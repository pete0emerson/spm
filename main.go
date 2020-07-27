package main

import (
	"github.com/pete0emerson/spm/pkg/spm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose int

var rootCmd = &cobra.Command{
	Use:              "spm",
	TraverseChildren: true,
}

func setVerbose() {
	if verbose == 0 {
		log.SetLevel(log.FatalLevel)
	} else if verbose == 1 {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {

	var installCommand = &cobra.Command{
		Use:   "install",
		Short: "Install package",
		Long:  "Install package",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			setVerbose()
			err := spm.Install(args[0], args[1], true)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	var removeCommand = &cobra.Command{
		Use:   "remove",
		Short: "Remove package",
		Long:  "Remove package",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			setVerbose()
			err := spm.Remove(args[0], false)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.AddCommand(installCommand)
	rootCmd.AddCommand(removeCommand)
	rootCmd.Flags().CountVarP(&verbose, "verbose", "v", "verbose output (use -vv to increase verbosity)")
	rootCmd.Execute()

}
