package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type RunArgs struct {
	ProjectName    string   `validate:"required,varname"`
	Hosts          []string `validate:"required"`
	NProcPerNode   string   `validate:"required"`
	ExperimentName string   `validate:"required,varname"`
	Port           int      `validate:"required,min=1"`
	RunName        string   `validate:"required,varname"`
	MaxRepeats     int      `validate:"required,min=-1"`
	Rest           []string
	ContainerName  *string
}

const runScript = `#!/usr/bin/env python
from higgsfield.internal.main import cli;
cli()
`

func nameFromRunArgs(args RunArgs) string {
	if args.ContainerName != nil && *args.ContainerName != "" {
		return *args.ContainerName
	}

	return DefaultProjExpContainerName(args.ProjectName, args.ExperimentName)
}

func trimPathForLength(path string, length int) string {
	// check if path is less than length
	if len(path) < length {
		return path
	}

	// get rid of home directory and replace is with ~
	// e.g. /home/user/... -> ~/...
	if path[0] == '/' {
		path = path[1:]
	}

	branches := strings.Split(path, "/")
	slashes := len(branches) - 1
	if slashes == 0 {
		return path[:length]
	}

	if branches[0] == "home" {
		path = "~/" + strings.Join(branches[2:], "/")
	}

	if len(path) < length {
		return path
	}

	return path[:length] + "..."
}

type nProcPerNode map[string][]int

// parseNProcPerNode converts
func parseNProcPerNode(host, nppn string) []int {
	var procMap nProcPerNode
	if err := json.Unmarshal([]byte(nppn), &procMap); err != nil {
		fmt.Printf("failed to parse nProcPerNode: %v\n", err)
		os.Exit(1)
	}

	hostNProc, ok := procMap[host]
	if !ok {
		fmt.Printf("failed to find host %s in nProcPerNode map\n", host)
		os.Exit(1)
	}

	return hostNProc
}

func Run(args RunArgs) {
	if err := Validator().Struct(args); err != nil {
		panic(err)
	}

	myIP, err := myPublicIP()
	if err != nil {
		fmt.Printf("failed to get my public IP: %v\n", err)
		os.Exit(1)
	}

	gpuIDs := parseNProcPerNode(myIP, args.NProcPerNode)

	master := args.Hosts[0]
	rank := 0

	if len(args.Hosts) > 1 {
		master, rank = rankAndMasterElseExit(args.Hosts)
	} else {
		master = "localhost"
	}

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

	containerName := nameFromRunArgs(args)

	fmt.Printf(`
╔══════════════════════════════════════════════════════════════════════════════════════════════════════
║  
║  > Training info:
║  > 🛠🛠🛠
║    
║  > EXPERIMENT NAME  = %s 
║  > RUN NAME         = %s
║  > CONTAINER NAME   = %s
║  > MODEL CHKPT PATH = %s
║  > GPU IDs          = %v
╚══════════════════════════════════════════════════════════════════════════════════════════════════════
`,
		args.ExperimentName,
		args.RunName,
		containerName,
		trimPathForLength(checkpointDir, 70),
		gpuIDs,
	)

	cmd, cmdArgs := buildArgs(
		nodeNum,
		rank,
		master,
		args.Port,
		[]string{"hf.py", "run"},
		len(gpuIDs),
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
	if err != nil {
		fmt.Printf("failed to create a file: %v\n", err)
	}
	defer f.Close()

	f.Write([]byte(runScript))

	dr := NewDockerRun(
		context.Background(),
		args.ProjectName,
		cwd,
		hostCachePath)
	if err := dr.Run(
		containerName,
		cmd,
		cmdArgs,
		args.Port,
		toStringSlice(gpuIDs),
	); err != nil {
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
		"--nproc_per_node",
		fmt.Sprint(nProcPerNode),
	}

	if master != "localhost" {
		args = append(args,
			"--master_addr",
			master,
			"--master_port",
			fmt.Sprint(masterPort),
		)
	}
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
