package net

import (
	"github.com/sirupsen/logrus"
	"github.com/xmapst/lightsocks/internal/constant"
	"github.com/xmapst/lightsocks/internal/protocol"
	"io"
	"net"
	"sync"
	"time"
)

type Relay struct {
	Src      net.Conn
	Dest     net.Conn
	Metadata *constant.Metadata
	Token    []byte
}

func (r *Relay) Start(s int) {
	switch s {
	case constant.Proxy:
		r.forward()
	default:
		r.direct()
	}
}

func (r *Relay) direct() {
	logrus.Infoln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, "accepted")
	defer func(src, dest net.Conn) {
		_ = dest.Close()
		_ = src.Close()
	}(r.Src, r.Dest)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(WriteOnlyWriter{Writer: r.Src}, ReadOnlyReader{Reader: r.Dest})
		_ = r.Src.SetReadDeadline(time.Now())
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(WriteOnlyWriter{Writer: r.Dest}, ReadOnlyReader{Reader: r.Src})
		_ = r.Dest.SetReadDeadline(time.Now())
	}()
	wg.Wait()
}

func (r *Relay) forward() {
	logrus.Infoln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, "accepted")
	defer func(src, dest net.Conn) {
		_ = dest.Close()
		_ = src.Close()
	}(r.Src, r.Dest)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		// dest --> encode --> src
		conn := &SecureTCPConn{
			ReadWriteCloser: r.Dest,
		}
		_ = conn.EncodeCopy(r.Token, r.Src)
		_ = r.Src.SetReadDeadline(time.Now())
	}()
	go func() {
		defer wg.Done()
		src := ReadOnlyReader{Reader: r.Src}
		dest := WriteOnlyWriter{Writer: r.Dest}
		// src --> decode --> dest
		for {
			pack, err := protocol.ReadFull(r.Token, src)
			if err != nil {
				break
			}
			logrus.Debugln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, pack.RandNu)
			_, err = dest.Write(pack.Payload)
			if err != nil {
				break
			}
		}
		_ = r.Dest.SetReadDeadline(time.Now())
	}()
	wg.Wait()
}
