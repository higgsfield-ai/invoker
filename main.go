package main

import (
	"fmt"
	"os"

	"github.com/ml-doom/invoker/internal/cli"
)

func main() {
	if err := cli.Cmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
