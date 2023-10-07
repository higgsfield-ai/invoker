package main

import (
	"fmt"
	"github.com/ml-doom/invoker/internal"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{Use: "higgsfield"}

var experimentCmd = &cobra.Command{Use: "experiment", Short: "Experiment commands"}

func runCmdFunc() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run an experiment",
		Run: func(cmd *cobra.Command, args []string) {
			internal.Run(internal.RunArgs{
				ExperimentName: internal.ParseOrExit[string](cmd, "experiment_name"),
				ProjectName:    internal.ParseOrExit[string](cmd, "project_name"),
				Port:           internal.ParseOrExit[int](cmd, "port"),
				RunName:        internal.ParseOrExit[string](cmd, "run_name"),
				NProcPerNode:   internal.ParseOrExit[int](cmd, "nproc_per_node"),
				Hosts:          internal.ParseOrExit[[]string](cmd, "hosts"),
				MaxRepeats:     -1,
				Rest:           args,
			})
		},
	}

	cmd.PersistentFlags().String("experiment_name", "", "name of the experiment")
	cmd.PersistentFlags().String("project_name", "", "name of the project")
	cmd.PersistentFlags().Int("port", 1234, "port to run the experiment on")
	cmd.PersistentFlags().String("run_name", "", "name of the run")
	cmd.PersistentFlags().Int("nproc_per_node", 1, "number of processes per node")
	cmd.PersistentFlags().StringSlice("hosts", []string{}, "list of hosts to run the experiment on")

	return cmd
}

func killCmdFunc() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill an experiment",
		Run: func(cmd *cobra.Command, args []string) {
			internal.Kill(internal.KillArgs{
				ProjectName:    internal.ParseOrExit[string](cmd, "project_name"),
				Hosts:          internal.ParseOrExit[[]string](cmd, "hosts"),
				ExperimentName: internal.ParseOrExit[string](cmd, "experiment_name"),
			})
		},
	}

	cmd.PersistentFlags().String("experiment_name", "", "name of the experiment")
	cmd.PersistentFlags().StringSlice("hosts", []string{}, "list of hosts to run the experiment on")
	cmd.PersistentFlags().String("project_name", "", "name of the project")

	return cmd
}

func decodeSecrets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode-secrets",
		Short: "Decode secrets",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			internal.DecodeSecrets(args[0])
		},
	}
	return cmd

}

func main() {
	experimentCmd.AddCommand(runCmdFunc())
	experimentCmd.AddCommand(killCmdFunc())

	rootCmd.AddCommand(decodeSecrets())
	rootCmd.AddCommand(experimentCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
