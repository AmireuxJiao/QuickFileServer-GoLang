package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	port int // 端口 flag
	dir  string

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "start file server",

		// args []string: file-server-cli server -p 123 /tmp --> [0]:[/tmp]
		PreRun: func(cmd *cobra.Command, args []string) {
			if port < 0 || port > 65535 {
				logrus.Fatal("端口配置错误")
				os.Exit(1)
			}
		},

		RunE: runServe,
	}
)

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8000, "监听端口")
	serveCmd.Flags().StringVarP(&dir, "dir", "d", ".", "文件服务器路径")
	rootCmd.AddCommand(serveCmd)
	logrus.Debug("add [server] command successfully")
}

func runServe(cmd *cobra.Command, args []string) error {
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	ipAddr := fmt.Sprintf("0.0.0.0:%d", port)
	logrus.Infof("Serving %s on http://%s", path, ipAddr)

	fs := http.FileServer(http.Dir(path))
	serve := &http.Server{Addr: ipAddr, Handler: fs}

	go func() {
		if err := serve.ListenAndServe(); err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("shutting down ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return serve.Shutdown(ctx)
}
