package cmd

import (
	"net"
	"strconv"

	"github.com/pkg/errors"
)

func isValidIP(ip string) (n netAddr, err error) {
	var host, port string
	if host, port, err = net.SplitHostPort(ip); err != nil {
		return netAddr{}, err
	}
	var p int
	if p, err = strconv.Atoi(port); err != nil {
		return netAddr{}, errors.Errorf("port must be an integer: %s", ip)
	}
	n = netAddr{host, p}
	return
}
