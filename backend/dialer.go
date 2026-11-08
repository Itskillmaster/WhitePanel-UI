package main

import (
	"net"
	"time"
)

func netDialer(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{Timeout: timeout}
	return d.Dial(network, address)
}
