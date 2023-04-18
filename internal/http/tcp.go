package http

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/lightsocks/internal/auth"
	"github.com/xmapst/lightsocks/internal/constant"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Proxy struct {
	wg   *sync.WaitGroup
	id   uuid.UUID
	conn net.Conn
	auth auth.Authenticator
}

func (p *Proxy) init(wg *sync.WaitGroup, id uuid.UUID, conn net.Conn) {
	p.wg = wg
	p.id = id
	p.conn = conn
	p.auth = new(auth.Auth)
}

func (p *Proxy) srcAddr() string {
	return p.conn.RemoteAddr().String()
}

func (p *Proxy) Handle(wg *sync.WaitGroup, id uuid.UUID, conn net.Conn, tcpIn chan<- *constant.TCPContext) error {
	p.init(wg, id, conn)
	lines, err := p.readString("\r\n")
	if err != nil {
		logrus.Errorln(p.id, p.srcAddr(), err)
		return err
	}
	if len(lines) < 2 {
		logrus.Errorln(p.id, p.srcAddr(), "request line error")
		return errors.New("request line error")
	}
	err = p.handshake(lines)
	if err != nil {
		logrus.Errorln(p.id, p.srcAddr(), err)
		return err
	}
	return p.processRequest(lines, tcpIn)
}

func (p *Proxy) handshake(lines []string) (err error) {
	var user, pass string
	for _, line := range lines {
		// get username/password
		if strings.HasPrefix(line, ProxyAuthorization) {
			line = strings.TrimPrefix(line, ProxyAuthorization)
			bs, err := base64.StdEncoding.DecodeString(line)
			if err != nil {
				logrus.Errorln(p.id, p.srcAddr(), err)
				continue
			}
			if bs == nil {
				continue
			}
			_auth := bytes.Split(bs, []byte(":"))
			if len(_auth) < 2 {
				continue
			}
			user, pass = string(_auth[0]), string(bytes.Join(_auth[1:], []byte(":")))
		}
	}
	if user != "" {
		logrus.Infoln(p.id, p.conn.RemoteAddr(), user)
	}
	// check username/password
	if p.auth.Enable() && !p.auth.Verify(user, pass, p.conn.RemoteAddr().String()) {
		logrus.Errorln(p.id, p.srcAddr(), "authentication failed")
		_, err = p.conn.Write([]byte{0x00, 0xff})
		if err != nil {
			logrus.Errorln(p.id, p.srcAddr(), err)
		}
		return errors.New("authentication failed")
	}
	return nil
}

func (p *Proxy) processRequest(lines []string, tcpIn chan<- *constant.TCPContext) error {
	requestLine := strings.Split(lines[0], " ")
	if len(requestLine) < 3 {
		logrus.Errorln(p.id, p.srcAddr(), "request line error")
		return errors.New("request line error")
	}
	method := requestLine[0]
	requestTarget := requestLine[1]
	version := requestLine[2]
	var err error
	if method == HTTPCONNECT {
		shp := strings.Split(requestTarget, ":")
		addr := shp[0]
		port, _ := strconv.Atoi(shp[1])
		err = p.handleHTTPConnectMethod(addr, uint16(port), tcpIn)
	} else {
		si := strings.Index(requestTarget, "//")
		restUrl := requestTarget[si+2:]
		if restUrl == "" {
			_, _ = p.conn.Write([]byte("HTTP/1.0 404 Not Found\r\n\r\n"))
			logrus.Errorln(p.id, p.srcAddr(), "404 Not Found")
			return errors.New("HTTP/1.0 404 Not Found")
		}
		port := 80
		ei := strings.Index(restUrl, "/")
		url := "/"
		hostPort := restUrl
		if ei != -1 {
			hostPort = restUrl[:ei]
			url = restUrl[ei:]
		}
		as := strings.Split(hostPort, ":")
		addr := as[0]
		if len(as) == 2 {
			port, _ = strconv.Atoi(as[1])
		}
		var header string
		for _, line := range lines[1:] {
			if strings.HasPrefix(line, ProxyAuthorization) {
				continue
			}
			if strings.HasPrefix(line, "Proxy-") {
				line = strings.TrimPrefix(line, "Proxy-")
			}
			header += fmt.Sprintf("%s\r\n", line)
		}
		newline := method + " " + url + " " + version + "\r\n" + header
		err = p.handleHTTPProxy(addr, uint16(port), newline, tcpIn)
	}
	return err
}

func (p *Proxy) httpWriteProxyHeader() {
	_, err := p.conn.Write([]byte("HTTP/1.1 200 OK Connection Established\r\n"))
	if err != nil {
		logrus.Warningln(p.id, p.srcAddr(), err)
		return
	}

	_, err = p.conn.Write([]byte(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123))))
	if err != nil {
		logrus.Warningln(p.id, p.srcAddr(), err)
		return
	}
	_, err = p.conn.Write([]byte("Transfer-Encoding: chunked\r\n"))
	if err != nil {
		logrus.Warningln(p.id, p.srcAddr(), err)
		return
	}
	_, err = p.conn.Write([]byte("\r\n"))
	if err != nil {
		logrus.Warningln(p.id, p.srcAddr(), err)
		return
	}
}

func (p *Proxy) handleHTTPConnectMethod(addr string, port uint16, tcpIn chan<- *constant.TCPContext) error {
	target := fmt.Sprintf("%s:%d", addr, port)

	tcpIn <- &constant.TCPContext{
		Conn: p.conn,
		Metadata: &constant.Metadata{
			ID:      p.id,
			NetWork: constant.TCP,
			Type:    constant.HTTPCONNECT,
			Src:     p.srcAddr(),
			Dest:    target,
		},
		PreFn: p.httpWriteProxyHeader,
		PostFn: func() {
			p.wg.Done()
		},
	}
	return nil
}

// Subsequent request lines are full paths, some servers may have problems
func (p *Proxy) handleHTTPProxy(addr string, port uint16, line string, tcpIn chan<- *constant.TCPContext) error {
	target := fmt.Sprintf("%s:%d", addr, port)

	tcpIn <- &constant.TCPContext{
		Conn: p.conn,
		Metadata: &constant.Metadata{
			ID:      p.id,
			NetWork: constant.TCP,
			Type:    constant.HTTP,
			Src:     p.srcAddr(),
			Dest:    target,
		},
		Line: line,
		PostFn: func() {
			p.wg.Done()
		},
	}
	return nil
}

func (p *Proxy) readString(delim string) ([]string, error) {
	var buf = make([]byte, 4096)
	_, err := io.ReadAtLeast(p.conn, buf, 1)
	if err != nil && err != io.EOF {
		logrus.Errorln(p.id, p.srcAddr(), err)
		return nil, err
	}
	return strings.Split(string(buf), delim), nil
}
