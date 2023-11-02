package misc

import (
	"fmt"
	"math/rand"
	"os"

	ns "github.com/cakturk/go-netstat/netstat"
	"github.com/pkg/errors"
)

func DSN(path string) string {
	return fmt.Sprintf("file:%s?_journal=WAL&_sync=normal&_vacuum=1", path)
}

const alphabet = "QWERTYUIOPASDFGHJKLZXCVBNMqwertuiopasdfghjklzxcvbnm1234567890"

func RandXAPIKey() string {
	// generate a random 32 character string
	// to be used as the X-API-KEY header

	key := make([]byte, 32)

	for i := 0; i < 32; i++ {
		key[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(key)
}

func listen(s *ns.SockTabEntry) bool { return s.State == ns.Listen }

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

func KillByPID(pid int) error {
  p, err := os.FindProcess(pid)
  if err != nil {
    return errors.WithMessagef(err, "failed to find process with pid %d", pid)
  }

  if err := p.Kill(); err != nil {
    return errors.WithMessagef(err, "failed to kill process with pid %d", pid)
  }

  return nil
}
