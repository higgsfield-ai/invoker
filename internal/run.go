package internal

import (
	"context"
	"fmt"
	"os"
)

type RunArgs struct {
	ProjectName    string   `validate:"required,varname"`
	Hosts          []string `validate:"required"`
	NProcPerNode   int      `validate:"required,min=1"`
	ExperimentName string   `validate:"required,varname"`
	Port           int      `validate:"required,min=1"`
	RunName        string   `validate:"required,varname"`
	MaxRepeats     int      `validate:"required,min=-1"`
	Rest           []string
}

const runScript = `#!/usr/bin/env python
from higgsfield.internal.main import cli;
cli()
`

func Run(args RunArgs) {
	if err := Validator().Struct(args); err != nil {
		panic(err)
	}

	master, rank := getRankAndMasterElseExit(args.Hosts)
	portIsAvailable(args.Port)
	nodeNum := len(args.Hosts)

	if !isPortAvailable(args.Port) {
		fmt.Printf("port %d is not available\n", args.Port)
		os.Exit(1)
	}

	hostCachePath, checkpointDir, err := makeDefaultDirectories(args.ProjectName, args.ExperimentName, args.RunName)
	if err != nil {
		fmt.Printf("failed to create directories: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(`
|=============================================================================================================
|=
|=  Training info:
|=  🛠🛠🛠
|=
|=  EXPERIMENT NAME =       %s
|=  RUN NAME =              %s
|=  MODEL CHECKPOINT PATH = %s
|=
|=============================================================================================================
`, args.ExperimentName, args.RunName, checkpointDir)

	cmd, cmdArgs := buildArgs(
		nodeNum,
		rank,
		master,
		args.Port,
		[]string{"hf.py", "run"},
		args.NProcPerNode,
		args.ExperimentName,
		args.RunName,
		args.MaxRepeats,
		args.Rest,
	)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("failed to get current working directory: %v\n", err)
		os.Exit(1)
	}

  // create a "higgsfield" file in cwd
  f, err := os.Create("hf.py")
  if err != nil  {
    fmt.Printf("failed to create a file: %v\n", err)
  }
  defer f.Close()

  f.Write([]byte(runScript))

	dr := NewDockerRun(context.Background(), args.ProjectName, cwd, hostCachePath)
	if err := dr.Run(args.ExperimentName, cmd, cmdArgs, args.Port); err != nil {
		fmt.Printf("error occured while running experiment: %+v\n", err)
		os.Exit(1)
	}
}

func buildArgs(
	nodeNum int,
	rank int,
	master string,
	masterPort int,
	experimentExecutable []string,
	nProcPerNode int,
	experimentName string,
	runName string,
	maxRepeats int,
	rest []string,
) (string, []string) {
	args := []string{
		"--nnodes",
		fmt.Sprint(nodeNum),
		"--node_rank",
		fmt.Sprint(rank),
		"--master_addr",
		master,
		"--master_port",
		fmt.Sprint(masterPort),
		"--nproc_per_node",
		fmt.Sprint(nProcPerNode)}

	args = append(args, experimentExecutable...)
	args = append(args,
		"--experiment_name",
		experimentName,
		"--run_name",
		runName,
		"--max_repeats",
		fmt.Sprint(maxRepeats))

	args = append(args, rest...)

	return "torchrun", args
}
