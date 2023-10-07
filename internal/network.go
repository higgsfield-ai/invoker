package internal

import (
	"fmt"
	"math/rand"
	"net"
)

func isPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	defer listener.Close()

	return true
}

func GeneratePort() int {
	port := rand.Intn(65535-1024) + 1024
	for !isPortAvailable(port) {
		port = rand.Intn(65535-1024) + 1024
	}

	return port
}
