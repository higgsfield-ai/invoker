package misc

import (
	"os"

	"github.com/pkg/errors"

	"path/filepath"
)

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

func MakeDefaultDirectories(projectName, experimentName, runName string) (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", errors.WithMessage(err, "failed to get user home directory")
	}

	cacheDir := Path{path: filepath.Join(home, ".cache")}
	if err = cacheDir.mkdirIfNotExists(); err != nil {
		return "", "", errors.WithMessage(err, "failed to create cache directory")
	}

	checkpointDir := cacheDir.Join("higgsfield").Join(projectName).Join("experiments").Join(experimentName).Join(runName)
	if err = checkpointDir.mkdirIfNotExists(); err != nil {
		return "", "", errors.WithMessagef(err, "failed to create checkpoint directory for experiment %s and run name %s", experimentName, runName)
	}

	return cacheDir.path, checkpointDir.path, nil
}

func MakeHiggsfieldDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.WithMessagef(err, "cannot find user home directory")
	}

	homeDir += "/higgsfield"

	if err := os.MkdirAll(homeDir, os.ModePerm); err != nil {
		return "", errors.WithMessagef(err, "failed to create higgsfield directory")
	}

	return homeDir, nil
}

func SetupCfgHFDir(projectName string) (cfgpath string, hfDir string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", errors.WithMessagef(err, "cannot find user home directory")
	}

	cfgpath = homeDir + "/higgsfield/setup.cfg"
	hfDir = homeDir + "/higgsfield/"

  if err := os.MkdirAll(hfDir, os.ModePerm); err != nil {
    return "", "", errors.WithMessagef(err, "cannot create higgsfield projects directory")
  }

	return cfgpath, hfDir, nil
}
