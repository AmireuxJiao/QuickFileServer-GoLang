package cmd

import (
	"context"
	"encoding/json"
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
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start file server",

	// args []string: file-server-cli server -p 123 /tmp --> [0]:[/tmp]
	PreRun: func(cmd *cobra.Command, args []string) {
		if port < 0 || port > 65535 {
			logrus.WithFields(logrus.Fields{
				"port": port,
			}).Fatal("端口配置错误")
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
	logrus.WithFields(logrus.Fields{
		"addrs": addrs,
	}).Debug("获取网络地址")

	var ips = getAllAccessibleAddr(addrs)
	for _, ip := range ips {
		logrus.WithFields(logrus.Fields{
			"ip":   ip,
			"port": port,
		}).Infof("服务地址")

		if qr != "false" {
			smallQr := qr == "small"
			if err := generateQRCode(ip, port, smallQr); err != nil {
				logrus.WithFields(logrus.Fields{
					"ip":    ip,
					"port":  port,
					"error": err,
				}).Warnf("生成二维码失败")
			}
		}
	}

	ipAddr := fmt.Sprintf("0.0.0.0:%d", port)
	logrus.WithFields(logrus.Fields{
		"path": path,
		"addr": ipAddr,
	}).Infof("开始提供服务")

	fs := http.FileServer(http.Dir(path))

	mux := http.NewServeMux()
	mux.Handle("/", fs)

	mux.HandleFunc("/list", listFiles)
	mux.HandleFunc("/ping", ping)
	mux.HandleFunc("/health", healthCheck)

	serve := &http.Server{Addr: ipAddr, Handler: mux}

	go func() {
		if err := serve.ListenAndServe(); err != http.ErrServerClosed {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatalf("监听失败")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("接收到退出信号，正在关闭服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serve.Shutdown(ctx); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warnf("关闭服务失败")
	} else {
		logrus.Info("服务已成功关闭")
	}

	return nil
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
		logrus.WithFields(logrus.Fields{
			"ip":    ip,
			"port":  port,
			"error": err,
		}).Fatalf("生成二维码失败")
	}

	if smallQr {
		fmt.Print(qrCode.ToSmallString(true))
	} else {
		fmt.Print(qrCode.ToString(true))
	}

	return nil
}

// 新增的 API 处理函数
func listFiles(w http.ResponseWriter, r *http.Request) {
	// http.Error(w, "Not implemented", http.StatusNotImplemented)
	files, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var fileList []string
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileList)
}

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "PONG")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}
