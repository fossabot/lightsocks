package resolver

import (
	"context"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"strings"
	"time"
)

var (
	// DefaultResolver aim to resolve ip
	DefaultResolver Resolver
	// DefaultDNSTimeout defined the default dns request timeout
	DefaultDNSTimeout = time.Second * 5
)

var (
	ErrIPNotFound = errors.New("couldn't find ip")
	ErrIPVersion  = errors.New("ip version error")
)

type Resolver interface {
	LookupIP(ctx context.Context, host string) ([]net.IP, error)
	LookupIPv4(ctx context.Context, host string) ([]net.IP, error)
	LookupIPv6(ctx context.Context, host string) ([]net.IP, error)
	ResolveIP(host string) (ip net.IP, err error)
	ResolveIPv4(host string) (ip net.IP, err error)
	ResolveIPv6(host string) (ip net.IP, err error)
	ExchangeContext(ctx context.Context, m *dns.Msg) (msg *dns.Msg, err error)
}

// LookupIPv4 with a host, return ipv4 list
func LookupIPv4(ctx context.Context, host string) ([]net.IP, error) {
	ip := net.ParseIP(host)
	if ip != nil {
		if !strings.Contains(host, ":") {
			return []net.IP{ip}, nil
		}
		return nil, ErrIPVersion
	}

	if DefaultResolver != nil {
		return DefaultResolver.LookupIPv4(ctx, host)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultDNSTimeout)
	defer cancel()
	ipAddrs, err := net.DefaultResolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return nil, err
	} else if len(ipAddrs) == 0 {
		return nil, ErrIPNotFound
	}

	return ipAddrs, nil
}

// ResolveIPv4 with a host, return ipv4
func ResolveIPv4(host string) (net.IP, error) {
	ips, err := LookupIPv4(context.Background(), host)
	if err != nil {
		return nil, err
	} else if len(ips) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrIPNotFound, host)
	}
	return ips[rand.Intn(len(ips))], nil
}

// LookupIPv6 with a host, return ipv6 list
func LookupIPv6(ctx context.Context, host string) ([]net.IP, error) {
	ip := net.ParseIP(host)
	if ip != nil {
		if strings.Contains(host, ":") {
			return []net.IP{ip}, nil
		}
		return nil, ErrIPVersion
	}

	if DefaultResolver != nil {
		return DefaultResolver.LookupIPv6(ctx, host)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultDNSTimeout)
	defer cancel()
	ipAddrs, err := net.DefaultResolver.LookupIP(ctx, "ip6", host)
	if err != nil {
		return nil, err
	} else if len(ipAddrs) == 0 {
		return nil, ErrIPNotFound
	}

	return ipAddrs, nil
}

// ResolveIPv6 with a host, return ipv6
func ResolveIPv6(host string) (net.IP, error) {
	ips, err := LookupIPv6(context.Background(), host)
	if err != nil {
		return nil, err
	} else if len(ips) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrIPNotFound, host)
	}
	return ips[rand.Intn(len(ips))], nil
}

// LookupIPWithResolver same as ResolveIP, but with a resolver
func LookupIPWithResolver(ctx context.Context, host string, r Resolver) ([]net.IP, error) {
	if r != nil {
		return r.LookupIP(ctx, host)
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return []net.IP{ip}, nil
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	} else if len(ips) == 0 {
		return nil, ErrIPNotFound
	}

	return ips, nil
}

// ResolveIP with a host, return ip
func LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	return LookupIPWithResolver(ctx, host, DefaultResolver)
}

// ResolveIP with a host, return ip
func ResolveIP(host string) (net.IP, error) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		logrus.Debugln(host, "time-consuming dns query", elapsed)
	}()
	ips, err := LookupIP(context.Background(), host)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	} else if len(ips) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrIPNotFound, host)
	}
	return ips[rand.Intn(len(ips))], nil
}