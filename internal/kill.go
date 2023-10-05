package internal

import (
	"context"
	"fmt"
	"os"
)

type KillArgs struct {
	ProjectName    string   `validate:"required,varname"`
	Hosts          []string `validate:"required,min=1"`
	ExperimentName string   `validate:"varname"`
}

func Kill(args KillArgs){ 
  if err := Validator().Struct(args); err != nil {
    panic(err)
  }

	getRankAndMasterElseExit(args.Hosts)

	// get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	cachePath := home + "/.cache/" + args.ProjectName + "/"
	metadataManager, err := NewMetadataManager(cachePath + "metadata.json")
	if err != nil {
		panic(err)
	}

	toKillContainers, err := metadataManager.KillExperiments(&args.ExperimentName)
	if err != nil {
		panic(err)
	}

	// get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dr := NewDockerRun(context.Background(), args.ProjectName, cwd, cachePath)

	for i := range toKillContainers {
		if err := dr.Kill(toKillContainers[i]); err != nil {
			fmt.Printf("error occured while killing containers: %+v\n", err)
		}
	}
}
