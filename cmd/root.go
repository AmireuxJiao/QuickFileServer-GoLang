package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const cliName = "flie-server-cli"

var (
	logLevel string // 全局 flag：日志级别

	rootCmd = &cobra.Command{
		Use:   cliName,
		Short: "File Share Server",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 在任何一个子命令执行前，统一初始化 logrus
			level, err := logrus.ParseLevel(logLevel)
			if err != nil {
				logrus.Fatalf("invalid log level: %v", err)
			}
			logrus.SetLevel(level)
			logrus.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			})
			logrus.SetOutput(os.Stdout)
		},
	}
)

func init() {
	// PersistentFlags()，作用范围：当前命令 + 所有子命令，常用于命令行的全局配置
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "debug",
		"日志级别: debug, info, warn, error, fatal, panic")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
