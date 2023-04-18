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
    start := time.Now()
    logrus.Infoln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, "accepted")
    defer func(src, dest net.Conn) {
        _ = dest.Close()
        _ = src.Close()
        logrus.Infoln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, "finish", time.Since(start))
    }(r.Src, r.Dest)
    wg := new(sync.WaitGroup)
    wg.Add(2)
    go func() {
        defer wg.Done()
        _, err := io.Copy(r.Src, r.Dest)
        if err != nil {
            logrus.Warningln(r.Metadata.ID, r.Metadata.Dest, "-->", r.Metadata.Src, err)
        }
        _ = r.Src.SetReadDeadline(time.Now())
    }()
    go func() {
        defer wg.Done()
        _, err := io.Copy(r.Dest, r.Src)
        if err != nil {
            logrus.Warningln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, err)
        }
        _ = r.Dest.SetReadDeadline(time.Now())
    }()
    wg.Wait()
}

func (r *Relay) forward() {
    start := time.Now()
    logrus.Infoln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, "accepted")
    defer func(src, dest net.Conn) {
        _ = dest.Close()
        _ = src.Close()
        logrus.Infoln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, "finish", time.Since(start))
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
        // src --> decode --> dest
        for {
            pack, err := protocol.ReadFull(r.Token, r.Src)
            if err != nil {
                break
            }
            logrus.Debugln(r.Metadata.ID, r.Metadata.Src, "-->", r.Metadata.Dest, pack.RandNu)
            _, err = r.Dest.Write(pack.Payload)
            if err != nil {
                break
            }
        }
        _ = r.Dest.SetReadDeadline(time.Now())
    }()
    wg.Wait()
}
