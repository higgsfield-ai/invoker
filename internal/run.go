package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type RunArgs struct {
	ProjectName    string   `validate:"required,varname"`
	Hosts          []string `validate:"required,min=1"`
	NProcPerNode   int      `validate:"required,min=1"`
	ExperimentName string   `validate:"required,varname"`
	Port           int      `validate:"required,min=1"`
	RunName        string   `validate:"required,varname"`
	MaxRepeats     int      `validate:"required,min=-1"`
	Rest           []string
}

func Run(args RunArgs) {
	if err := Validator().Struct(args); err != nil {
		panic(err)
	}

	master, rank := getRankAndMasterElseExit(args.Hosts)
	portIsAvailable(args.Port)
	nodeNum := len(args.Hosts)

	hostCachePath, guestLogPath, err := makeDefaultDirectories(args.ProjectName, args.ExperimentName, args.RunName)
	if err != nil {
		fmt.Printf("failed to create directories: %v\n", err)
		os.Exit(1)
	}

	mgr, err := NewMetadataManager(filepath.Join(hostCachePath, args.ProjectName, "metadata.json"))
	if err != nil {
		fmt.Printf("failed to create metadata manager: %v\n", err)
		os.Exit(1)
	}

	if _, err = mgr.KillExperiments(&args.ExperimentName); err != nil {
		fmt.Printf("failed to kill experiments: %v\n", err)
	}

	cmd, cmdArgs := buildArgs(
		nodeNum,
		rank,
		master,
		args.Port,
		guestLogPath,
		[]string{"run_experiment.py"},
		args.NProcPerNode,
		args.ExperimentName,
		args.Port,
		args.RunName,
		args.MaxRepeats,
		args.Rest,
	)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("failed to get current working directory: %v\n", err)
		os.Exit(1)
	}

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
	logPath string,
	experimentExecutable []string,
	nProcPerNode int,
	experimentName string,
	port int,
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
		fmt.Sprint(nodeNum),
		"--nproc_per_node",
		fmt.Sprint(nProcPerNode),
		"--log_dir",
		logPath}

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
