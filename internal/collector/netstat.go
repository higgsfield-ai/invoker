package collector

import (
	ns "github.com/cakturk/go-netstat/netstat"
	"github.com/pkg/errors"
)

func listen(s *ns.SockTabEntry) bool {
	return s.State == ns.Listen
}

func ProcOnPort(port uint16) (*ns.SockTabEntry, error) {
	tabs, err := ns.TCPSocks(func(s *ns.SockTabEntry) bool { return s.State == ns.Listen })
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get IPv4 TCP sockets for port %d", port)
	}

	tabs6, err := ns.TCP6Socks(listen)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get IPv6 TCP sockets for port %d", port)
	}

	tabs = append(tabs, tabs6...)

	for _, tab := range tabs {
		if tab.LocalAddr.Port == port {
			return &tab, nil
		}
	}

	return nil, nil
}
