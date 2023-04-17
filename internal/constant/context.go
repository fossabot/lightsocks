package constant

import (
	"net"
)

// Socks addr type
const (
	TCP NetWork = iota
	UDP
)

const (
	HTTP Type = iota
	HTTPCONNECT
	SOCKS4
	SOCKS5
)

const (
	Direct = iota
	Proxy
)

// TCPContext is used to store connection address
type TCPContext struct {
	Conn     net.Conn
	Metadata *Metadata
	Line     string // http proxy
	PreFn    func()
	PostFn   func()
}
