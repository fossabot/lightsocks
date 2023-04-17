package mixed

import (
	"github.com/gofrs/uuid"
	"github.com/xmapst/lightsocks/internal/constant"
	"github.com/xmapst/lightsocks/internal/http"
	"github.com/xmapst/lightsocks/internal/socks4"
	"github.com/xmapst/lightsocks/internal/socks5"
	"net"
	"sync"
)

type Proxy interface {
	Handle(wg *sync.WaitGroup, uuid uuid.UUID, conn net.Conn, tcpIn chan<- *constant.TCPContext) error
}

func newSocks4() Proxy {
	return &socks4.Proxy{}
}

func newSocks5(udpAddr string) Proxy {
	return &socks5.Proxy{Udp: udpAddr}
}

func newHttp() Proxy {
	return &http.Proxy{}
}
