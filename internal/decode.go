package internal

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func DecodeSecrets(secrets string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("failed to get current working directory: %v\n", err)
		os.Exit(1)
	}

	decoded, err := base64.StdEncoding.DecodeString(secrets)

	if err != nil {
		fmt.Printf("failed to decode base64 string: %v\n", err)
	}

	f, err := os.Create(filepath.Join(cwd, "env"))
	if err != nil {
		fmt.Printf("failed to create env file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	_, err = f.Write(decoded)
	if err != nil {
		fmt.Printf("failed to write to env file: %v\n", err)
		os.Exit(1)
	}
}
