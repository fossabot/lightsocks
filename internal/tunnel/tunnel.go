package tunnel

import (
	"crypto/tls"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/chanx"
	"github.com/xmapst/lightsocks/internal/conf"
	"github.com/xmapst/lightsocks/internal/constant"
	N "github.com/xmapst/lightsocks/internal/net"
	"github.com/xmapst/lightsocks/internal/resolver"
	"github.com/xmapst/lightsocks/internal/statistic"
	"net"
	"runtime"
	"strconv"
)

var (
	TCPIn   = chanx.NewUnboundedChan[*constant.TCPContext](10000)
	workers = 4
)

func Start(token string) {
	go process(token)
}

// processTCP starts a loop to handle tcp packet
func processTCP(token string) {
	for conn := range TCPIn.Out {
		go handleTCPConn(conn, token)
	}
}

func process(token string) {
	if num := runtime.GOMAXPROCS(0); num > workers {
		workers = num
	}
	workers *= workers
	for i := 0; i < workers; i++ {
		go processTCP(token)
	}
}

func handleTCPConn(ctx *constant.TCPContext, token string) {
	dial := net.Dialer{Timeout: conf.App.Timeout}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(ctx.Conn)

	// connect to the target
	target := ctx.Metadata.Dest
	if conf.App.Mode == conf.ClientMode {
		target = constant.IP{
			Addr: conf.App.Server.Host,
			Port: conf.App.Server.Port,
		}
	}
	ip, err := resolver.ResolveIP(target.Addr)
	if err != nil {
		logrus.Errorln(ctx.Metadata.ID, ctx.Metadata.Src, "-->", ctx.Metadata.Dest, err)
		return
	}
	destAddr := net.JoinHostPort(ip.String(), strconv.FormatInt(target.Port, 10))
	var destConn net.Conn
	if conf.App.TLS.Enable && conf.App.Mode == conf.ClientMode {
		destConn, err = tls.DialWithDialer(&dial, "tcp", destAddr, conf.App.TLSConf)
	} else {
		destConn, err = dial.Dial("tcp", destAddr)
	}
	if err != nil {
		logrus.Errorln(ctx.Metadata.ID, ctx.Metadata.Src, "-->", ctx.Metadata.Dest, err)
		return
	}
	defer func(destConn net.Conn) {
		_ = destConn.Close()
	}(destConn)

	// 发送被代理的信息
	if conf.App.Mode == conf.ClientMode {
		destSecConn := &N.SecureTCPConn{ReadWriteCloser: destConn}
		_, err = destSecConn.EncodeWrite([]byte(token), []byte(ctx.Metadata.Dest.String()))
		if err != nil {
			logrus.Errorln(ctx.Metadata.ID, ctx.Metadata.Src, "-->", ctx.Metadata.Dest, err)
			return
		}
		// redirect http proxy
		if ctx.Line != "" {
			_, err = destSecConn.EncodeWrite([]byte(token), []byte(ctx.Line))
			if err != nil {
				logrus.Errorln(ctx.Metadata.ID, ctx.Metadata.Src, "-->", ctx.Metadata.Dest, err)
				return
			}
		}
	}

	if ctx.PreFn != nil {
		ctx.PreFn()
	}
	defer func() {
		if ctx.PostFn != nil {
			ctx.PostFn()
		}
	}()
	// direct http proxy
	var src, dest, _type = ctx.Conn, destConn, constant.Direct
	if conf.App.Mode == conf.DirectMode {
		if ctx.Line != "" {
			_, err = destConn.Write([]byte(ctx.Line))
			if err != nil {
				logrus.Errorln(ctx.Metadata.ID, ctx.Metadata.Src, "-->", ctx.Metadata.Dest, err)
				return
			}
		}
	} else {
		_type = constant.Proxy
		if conf.App.Mode == conf.ClientMode {
			src, dest = destConn, ctx.Conn
		}
	}
	dest = statistic.NewTCPTracker(dest, ctx.Metadata)
	relay := &N.Relay{
		Src:      src,
		Dest:     dest,
		Metadata: ctx.Metadata,
		Token:    []byte(token),
	}
	relay.Start(_type)
}
