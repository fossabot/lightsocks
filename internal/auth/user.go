package auth

import (
	"github.com/xmapst/lightsocks/internal/conf"
	"net"
)

type Authenticator interface {
	Verify(user string, pass string, addr string) bool
	Enable() bool
}

type Auth struct{}

func (a *Auth) Verify(username, password, addr string) bool {
	var res conf.User
	var ok bool
	for _, v := range conf.App.Users {
		if v.UserName == username {
			ok = true
			res = v
		}
	}
	if !ok {
		return false
	}
	if password != "" {
		if res.Password != password {
			return false
		}
	}
	if res.CIDR == nil {
		return true
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	return verifyCIDR(host, res.CIDR)
}

func (a *Auth) Enable() bool {
	return conf.App.Users != nil
}
