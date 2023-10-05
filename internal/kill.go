package internal

import (
	"context"
	"os"
)

type KillArgs struct {
	ProjectName    string   `validate:"required,varname"`
	Hosts          []string `validate:"required,min=1"`
	ExperimentName string   `validate:"varname"`
}

func Kill(args KillArgs) {
	if err := Validator().Struct(args); err != nil {
		panic(err)
	}

	getRankAndMasterElseExit(args.Hosts)

	// get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	cachePath := home + "/.cache/" + args.ProjectName + "/" + "experiments/"

	// get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dr := NewDockerRun(context.Background(), args.ProjectName, cwd, cachePath)

  if err := dr.Kill(args.ExperimentName); err != nil {
    panic(err)
  }
}
