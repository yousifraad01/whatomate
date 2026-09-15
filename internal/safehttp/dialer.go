// Package safehttp provides an HTTP dialer that refuses connections to
// private, loopback and link-local addresses after DNS resolution. Every
// outbound request whose destination is configured by a tenant (webhooks,
// custom actions, chatbot API nodes, IVR callbacks, SSO providers) must go
// through it so the server cannot be used to reach internal services.
package safehttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Dialer returns a DialContext function that blocks connections to
// private/loopback IPs after DNS resolution.
func Dialer() func(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		ips, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("no addresses found for %s", host)
		}

		for _, ipStr := range ips {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				continue
			}
			if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
				ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
				return nil, fmt.Errorf("connection to private address %s is not allowed", ipStr)
			}
		}

		// Connect to first resolved IP
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0], port))
	}
}

// NewTransport returns an http.Transport that dials through Dialer with
// sensible pooling defaults.
func NewTransport() *http.Transport {
	return &http.Transport{
		DialContext:         Dialer(),
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
}
