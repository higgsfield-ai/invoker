package misc

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"os"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"

	"math/rand"
)

const url = "https://api.ipify.org"

func MyPublicIP() (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.WithMessage(err, "failed to get public IP")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", errors.WithMessage(err, "failed to read response body")
	}

	return string(body), nil
}

func GetRankAndMasterElseExit(hosts []string) (string, int) {
	ip, err := MyPublicIP()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	master, rank := hosts[0], -1
	for i, host := range hosts {
		if host == ip {
			rank = i
			break
		}
	}

	if rank == -1 {
		fmt.Printf("%s not found in hosts list, omitting\n", ip)
		os.Exit(0)
	}

	return master, rank
}

func PortIsAvailable(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("port %d is already in use\n", port)
		os.Exit(1)
	}

	defer listener.Close()
}

func IsPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	defer listener.Close()

	return true
}

var (
	isolatedPorts = mapset.NewSet(
		55555, // collector port
	)
)

func GeneratePort() int {
	port := rand.Intn(65535-1024) + 1024
	for !IsPortAvailable(port) && isolatedPorts.Contains(port) {
		port = rand.Intn(65535-1024) + 1024
	}

	return port
}
