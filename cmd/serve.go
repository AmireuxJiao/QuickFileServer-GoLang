package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var (
	port     int // 端口 flag
	dir      string
	qr       string
	rootPath string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start file server",
	Long: `Start a file server to share files in the specified directory via HTTP.
Examples:
  # Start server on port 8080, share current directory, and generate QR code
  file-server-cli serve -p 8080 -d . -q true

  # Start server with default port (9999) and share /tmp directory
  file-server-cli serve -d /tmp`,

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
	serveCmd.Flags().StringVarP(&qr, "qrcode", "q", "false",
		"生成URL二维码访问图 (values: 'false' (disable), 'true' (normal size), 'small' (compact size))")
	rootCmd.AddCommand(serveCmd)
	logrus.Debug("add [server] command successfully")
}

func runServe(cmd *cobra.Command, args []string) error {
	// path:当前文件服务器的工作地址
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	rootPath = path
	logrus.Debug("filepath.Abs(dir) : path = ", path)

	// net.InterfaceAddrs:
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
			if qrCode, err := generateQRCode(ip, port); err != nil {
				logrus.WithFields(logrus.Fields{
					"ip":    ip,
					"port":  port,
					"error": err,
				}).Warnf("生成二维码失败")
			} else { // 正常生成二维码
				if smallQr {
					fmt.Print(qrCode.ToSmallString(true))
				} else {
					fmt.Print(qrCode.ToString(true))
				}
			}
		}
	}

	r := gin.Default()
	// 修复后代码
	r.GET("/files/*filepath", preventPathTraversal(path), func(ctx *gin.Context) {
		realFilePath := strings.TrimPrefix(ctx.Request.URL.Path, "/files")
		if realFilePath == "" {
			realFilePath = "/"
		}
		ctx.Request.URL.Path = realFilePath
		http.FileServer(http.Dir(path)).ServeHTTP(ctx.Writer, ctx.Request)
	})

	r.GET("/list", listFiles)
	r.GET("/ping", ping)
	r.GET("/health", healthCheck)
	r.POST("/upload", uploadFile)

	ipAddr := fmt.Sprintf("0.0.0.0:%d", port)
	logrus.WithFields(logrus.Fields{
		"path": path,
		"addr": ipAddr,
	}).Infof("开始提供服务")

	serve := &http.Server{
		Addr:    ipAddr,
		Handler: r,
	}

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

func preventPathTraversal(rootDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户请求的相对路径（如 "subdir/file.txt" 或 "../../etc/passwd"）
		reqPath := c.Param("filepath") // 需在路由中定义占位符，如 "/files/*filepath"
		fullPath := filepath.Join(rootDir, reqPath)
		if !strings.HasPrefix(fullPath, rootDir) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid path"})
			return
		}
		c.Next()
	}
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

func generateQRCode(ip string, port int) (*qrcode.QRCode, error) {
	qrCode, err := qrcode.New(fmt.Sprintf("http://%s:%d/files", ip, port), qrcode.Medium)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ip":    ip,
			"port":  port,
			"error": err,
		}).Fatalf("生成二维码失败")
		return nil, err
	}

	return qrCode, nil
}

func listFiles(c *gin.Context) {
	files, err := os.ReadDir(dir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var fileList []string
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}
	c.JSON(http.StatusOK, fileList)
}

func ping(c *gin.Context) {
	c.String(http.StatusOK, "PONG")
}

func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func uploadFile(c *gin.Context) {
	c.Request.ParseMultipartForm(100 << 20) // 100MB
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "get file failed"})
	}
	defer file.Close()

	savePath := filepath.Join(rootPath, header.Filename)
	if !strings.HasPrefix(savePath, rootPath) {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid filename"})
		return
	}

	output, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create file failed: " + err.Error()})
		return
	}
	defer output.Close()
	io.Copy(output, file)

	c.JSON(http.StatusOK, gin.H{"message": "file uploaded successfully", "filename": header.Filename})
}
