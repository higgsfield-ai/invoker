package cli

import (
	"fmt"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/ml-doom/invoker/internal/collector"
	"github.com/ml-doom/invoker/internal/invoker"
	"github.com/ml-doom/invoker/internal/misc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "higgsfield"}

var experimentCmd = &cobra.Command{Use: "experiment", Short: "Experiment commands"}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run an experiment",
		Run: func(cmd *cobra.Command, args []string) {
			invoker.Run(invoker.RunArgs{
				ExperimentName: misc.ParseOrExit[string](cmd, "experiment_name"),
				ProjectName:    misc.ParseOrExit[string](cmd, "project_name"),
				Port:           misc.ParseOrExit[int](cmd, "port"),
				RunName:        misc.ParseOrExit[string](cmd, "run_name"),
				NProcPerNode:   misc.ParseOrExit[int](cmd, "nproc_per_node"),
				Hosts:          misc.ParseOrExit[[]string](cmd, "hosts"),
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

func killCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill an experiment",
		Run: func(cmd *cobra.Command, args []string) {
			invoker.Kill(invoker.KillArgs{
				ProjectName:    misc.ParseOrExit[string](cmd, "project_name"),
				Hosts:          misc.ParseOrExit[[]string](cmd, "hosts"),
				ExperimentName: func() string { v, _ := cmd.Flags().GetString("experiment_name"); return v }(),
			})
		},
	}

	cmd.PersistentFlags().String("experiment_name", "", "name of the experiment")
	cmd.PersistentFlags().StringSlice("hosts", []string{}, "list of hosts to run the experiment on")
	cmd.PersistentFlags().String("project_name", "", "name of the project")

	return cmd
}

func decodeSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode-secrets",
		Short: "Decode secrets",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			misc.DecodeSecrets(args[0])
		},
	}
	return cmd

}

func randomNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "random-name",
		Short: "Generate a random name",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(namesgenerator.GetRandomName(0))
		},
	}

	return cmd
}

func randomPortCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "random-port",
		Short: "Generate a random port",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(misc.GeneratePort())
		},
	}

	return cmd
}

func collectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collector",
		Short: "Run log collector",
		RunE: func(cmd *cobra.Command, args []string) error {
			server, err := collector.Initialize(collector.ServeArgs{
				ProjectName: misc.ParseOrExit[string](cmd, "project_name"),
				Port:        misc.ParseOrExit[int](cmd, "port"),
			})

      misc.ParseOrExit[bool](cmd, "force")
			if err != nil {
				if errors.Is(err, collector.ErrPortInUse) {
					return errors.WithMessagef(err, "cannot run collector, since other process is listening")
				}

				if !errors.Is(err, collector.ErrPortInUseByInvoker) {
					return errors.WithMessagef(err, "failed to initialize collector")
				}
			}

			if server != nil {
				fmt.Printf("collector listening on %s\n", server.Addr)
				if err := server.ListenAndServe(); err != nil {
					return errors.WithMessagef(err, "failed to listen and serve")
				}
			} else {
				fmt.Println("collector is already running")
			}

			return nil
		},
	}

	cmd.PersistentFlags().String("project_name", "", "name of the project")
	cmd.PersistentFlags().Int("port", 1234, "port to run collector on")
	cmd.PersistentFlags().Bool("force", false, "force collector to run on port even if it is already in use by some other process")
  cmd.PersistentFlags().Bool("daemon", false, "run collector in daemon mode")

	return cmd
}

func Cmd() *cobra.Command {
	experimentCmd.AddCommand(runCmd())
	experimentCmd.AddCommand(killCmd())

	rootCmd.AddCommand(decodeSecretsCmd())
	rootCmd.AddCommand(randomNameCmd())
	rootCmd.AddCommand(randomPortCmd())
	rootCmd.AddCommand(collectorCmd())
	rootCmd.AddCommand(experimentCmd)

	return rootCmd
}
