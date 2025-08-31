package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var (
	port int // 端口 flag
	dir  string
	qr   string
	// smallQr bool
)

var serveCmd = &cobra.Command{
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

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 9999, "监听端口")
	serveCmd.Flags().StringVarP(&dir, "dir", "d", ".", "文件服务器路径")
	serveCmd.Flags().StringVarP(&qr, "qrcode", "q", "false", "生成二维码 (true, false, small)")
	rootCmd.AddCommand(serveCmd)
	logrus.Debug("add [server] command successfully")
}

func runServe(cmd *cobra.Command, args []string) error {
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	logrus.Debug("  addrs = ", addrs)

	var ips = getAllAccessibleAddr(addrs)
	for _, ip := range ips {
		logrus.Infof("  http://%s:%d", ip, port)

		if qr != "false" {
			smallQr := qr == "small"
			if err := generateQRCode(ip, port, smallQr); err != nil {
				logrus.Warnf("failed to generate QR code: %v", err)
			}
		}
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

func getAllAccessibleAddr(addrs []net.Addr) []string {
	var ips []string

	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ips = append(ips, ipNet.IP.String())
		}
	}

	if len(ips) == 0 {
		ips = []string{"127.0.0.1"} // fallback
	}
	return ips
}

func generateQRCode(ip string, port int, smallQr bool) error {
	qrCode, err := qrcode.New(fmt.Sprintf("http://%s:%d", ip, port), qrcode.Medium)
	if err != nil {
		logrus.Fatalf("generate [QR] code error:%s", err)
	}

	if smallQr {
		fmt.Print(qrCode.ToSmallString(true))
	} else {
		fmt.Print(qrCode.ToString(true))
	}

	return nil
}
