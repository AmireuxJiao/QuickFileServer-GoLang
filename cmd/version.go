package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const Version = "0.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version ",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.WithField("version", Version).Debug("current version")
		fmt.Println(cliName, Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	logrus.Debug("add [version] command successfully")
}
