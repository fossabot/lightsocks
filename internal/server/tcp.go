package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/pires/go-proxyproto"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/lightsocks/internal/auth"
	"github.com/xmapst/lightsocks/internal/conf"
	"github.com/xmapst/lightsocks/internal/constant"
	N "github.com/xmapst/lightsocks/internal/net"
	"github.com/xmapst/lightsocks/internal/protocol"
	"github.com/xmapst/lightsocks/internal/tunnel"
	"net"
	"strconv"
	"sync"
	"time"
)

type Listener struct {
	tcp  net.Listener
	wg   *sync.WaitGroup
	conf *conf.Config
}

func (l *Listener) RawAddress() string {
	return fmt.Sprintf("%s:%d", l.conf.Local.Host, l.conf.Local.Port)
}

func (l *Listener) Address() string {
	return l.tcp.Addr().String()
}

func (l *Listener) close() {
	if l.tcp != nil {
		_ = l.tcp.Close()
		return
	}
	return
}

func (l *Listener) State() bool {
	return l.tcp != nil
}

func (l *Listener) Shutdown() error {
	return l.ShutdownWithTimeout(0)
}

func (l *Listener) ShutdownWithTimeout(timeout time.Duration) error {
	l.close()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c := make(chan struct{})
	go func() {
		defer close(c)
		l.wg.Wait()
	}()
	defer func() {
		logrus.Infoln("server closed")
		l.tcp = nil
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c:
		return nil
	}
}

func New() *Listener {
	return &Listener{
		wg:   new(sync.WaitGroup),
		conf: conf.App,
	}
}

func (l *Listener) ListenAndServe() (err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", l.RawAddress())
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	if l.conf.TLS.Enable {
		l.tcp, err = tls.Listen("tcp", tcpAddr.String(), l.conf.TLSConf)
	} else {
		l.tcp, err = net.ListenTCP("tcp", tcpAddr)
	}
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	logrus.Infoln("TCP Server Listening At:", l.tcp.Addr().String())
	ln := &proxyproto.Listener{Listener: l.tcp}
	tcpIn := tunnel.TCPIn.In
	for {
		var conn net.Conn
		conn, err = ln.Accept()
		if err != nil {
			continue
		}
		clientIP := conn.RemoteAddr().String()
		if !auth.VerifyIP(clientIP) {
			logrus.Warningln(clientIP, "access denied, not in allowed address group")
			_ = conn.Close()
		} else {
			l.wg.Add(1)
			bufConn := N.NewBufferedConn(conn)
			go l.handle(bufConn, tcpIn)
		}
	}
}

func (l *Listener) handle(srcConn net.Conn, tcpIn chan<- *constant.TCPContext) {
	id, _ := uuid.NewV4()
	packet, err := protocol.ReadFull([]byte(l.conf.Local.Token), srcConn)
	if err != nil {
		l.wg.Done()
		logrus.Errorln(id, srcConn.RemoteAddr(), err)
		_ = srcConn.Close()
		return
	}
	destAddr := string(packet.Payload)
	logrus.Debugln(id, srcConn.RemoteAddr(), "-->", destAddr, packet.RandNu)
	tcpIn <- &constant.TCPContext{
		Conn: srcConn,
		Metadata: &constant.Metadata{
			ID:      id,
			NetWork: constant.TCP,
			Type:    constant.SOCKS5,
			Src: func() constant.IP {
				host, port, _ := net.SplitHostPort(srcConn.RemoteAddr().String())
				_port, _ := strconv.ParseInt(port, 10, 64)
				return constant.IP{
					Addr: host,
					Port: _port,
				}
			}(),
			Dest: func() constant.IP {
				host, port, _ := net.SplitHostPort(destAddr)
				_port, _ := strconv.ParseInt(port, 10, 64)
				return constant.IP{
					Addr: host,
					Port: _port,
				}
			}(),
		},
		PostFn: func() {
			l.wg.Done()
		},
	}
}
