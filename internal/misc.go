package internal

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"path/filepath"
)

const url = "https://api.ipify.org"

func myPublicIP() (string, error) {
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

func getRankAndMasterElseExit(hosts []string) (string, int) {
	ip, err := myPublicIP()
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

func portIsAvailable(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("port %d is already in use\n", port)
		os.Exit(1)
	}

	defer listener.Close()
}

type Path struct {
	path string
}

func (p *Path) mkdirIfNotExists() error {
	// Check if the directory already exists
	_, err := os.Stat(p.path)
	if os.IsNotExist(err) {
		// Directory doesn't exist, so create it
		err := os.MkdirAll(p.path, os.ModePerm)
		if err != nil {
			return errors.WithMessagef(err, "failed to create directory %s", p.path)
		}
	} else if err != nil {
		return errors.WithMessagef(err, "failed to stat directory %s", p.path)
	}

	return nil
}

func (p *Path) Join(subpath string) *Path {
	return &Path{path: filepath.Join(p.path, subpath)}
}

func makeDefaultDirectories(projectName, experimentName, runName string) (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", errors.WithMessage(err, "failed to get user home directory")
	}

	cacheDir := Path{path: filepath.Join(home, ".cache")}
	if err = cacheDir.mkdirIfNotExists(); err != nil {
		return "", "", errors.WithMessage(err, "failed to create cache directory")
	}

	projectDir := cacheDir.Join(projectName)
	if err = projectDir.mkdirIfNotExists(); err != nil {
		return "", "", errors.WithMessage(err, "failed to create project directory")
	}

	experimentsDir := projectDir.Join("experiments")
	if err = experimentsDir.mkdirIfNotExists(); err != nil {
		return "", "", errors.WithMessage(err, "failed to create experiments directory")
	}

	experimentDir := experimentsDir.Join(experimentName)
	if err = experimentDir.mkdirIfNotExists(); err != nil {
		return "", "", errors.WithMessage(err, "failed to create experiment directory")
	}

	med := []string{"checkpoints", "sharded-checkpoints", "lr-schedules", "logs", "plots", "results"}
	for _, dir := range med {
		if err = experimentDir.Join(dir).Join(runName).mkdirIfNotExists(); err != nil {
			return "", "", errors.WithMessagef(err, "failed to create %s directory for experiment %s and run name %s", dir, experimentName, runName)
		}
	}

	return cacheDir.path, experimentDir.Join("logs").Join(runName).path, nil
}




func exitIfError(flag string, err error) {
	if err != nil {
		fmt.Printf("cannot parse %s: %v\n", flag, err)
		os.Exit(1)
	}

}

func ParseOrExit[T ~string| ~int | ~[]string](cmd *cobra.Command, flag string) T {
  got := parseOrExitInternal[T](cmd, flag)
  return got.(T)
}

func parseOrExitInternal[T ~string| ~int | ~[]string](cmd *cobra.Command, flag string) interface{} {
	var value T
	switch v := any(value).(type) {
	case string:
		v, err := cmd.Flags().GetString(flag)
    exitIfError(flag, err)
    return v
	case int:
		v, err := cmd.Flags().GetInt(flag)
    exitIfError(flag, err)
     return v
	case []string:
		v, err := cmd.Flags().GetStringSlice(flag)
    exitIfError(flag, err)
    return v
	default:
		fmt.Printf("cannot parse %s: unknown type %T\n", flag, v)
    os.Exit(1)
	}

  return nil
}
