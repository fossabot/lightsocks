package mixed

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/pires/go-proxyproto"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/lightsocks/internal/auth"
	"github.com/xmapst/lightsocks/internal/conf"
	"github.com/xmapst/lightsocks/internal/constant"
	N "github.com/xmapst/lightsocks/internal/net"
	"github.com/xmapst/lightsocks/internal/socks4"
	"github.com/xmapst/lightsocks/internal/socks5"
	"github.com/xmapst/lightsocks/internal/tunnel"
	"github.com/xmapst/lightsocks/internal/udp"
	"net"
	"sync"
	"time"
)

type Listener struct {
	tcp  net.Listener
	udp  *net.UDPConn
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
	if l.tcp != nil && l.udp != nil {
		_ = l.tcp.Close()
		_ = l.udp.Close()
		return
	}
	if l.tcp != nil {
		_ = l.tcp.Close()
		return
	}
	if l.udp != nil {
		_ = l.udp.Close()
		return
	}
	return
}

func (l *Listener) State() bool {
	return l.tcp != nil && l.udp != nil
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
		l.udp = nil
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
	l.tcp, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	udpAddr, err := net.ResolveUDPAddr("udp", l.RawAddress())
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	l.udp, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		logrus.Errorln(err)
		return err
	}

	// udp
	go func() {
		logrus.Infoln("UDP Server Listening At:", l.udp.LocalAddr().String())
		udp.Listen(l.udp)
	}()
	// tcp
	listenAddr := []string{
		fmt.Sprintf("http://%s", l.Address()),
		fmt.Sprintf("socks4://%s", l.Address()),
		fmt.Sprintf("socks5://%s", l.Address()),
	}
	logrus.Infoln("TCP Server Listening At:", listenAddr)
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
			go l.handle(conn, tcpIn)
		}
	}
}

func (l *Listener) handle(conn net.Conn, tcpIn chan<- *constant.TCPContext) {
	id, _ := uuid.NewV4()
	bufConn := N.NewBufferedConn(conn)
	head, err := bufConn.Peek(1)
	if err != nil {
		logrus.Errorln(id, err)
		_ = conn.Close()
		return
	}
	var proxy Proxy
	switch head[0] {
	case socks4.Version:
		proxy = newSocks4()
	case socks5.Version:
		proxy = newSocks5(l.udp.LocalAddr().String())
	default:
		proxy = newHttp()
	}
	err = proxy.Handle(l.wg, id, bufConn, tcpIn)
	if err != nil {
		_ = conn.Close()
	}
}
