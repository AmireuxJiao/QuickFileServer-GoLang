package cmd

import (
	"bytes"
	"fmt"
	"io" // 新增：导入io包，用于io.Copy
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	// 1. 保存原始环境（测试后恢复，避免影响其他用例）
	originalStdout := os.Stdout
	originalLogOutput := logrus.StandardLogger().Out
	originalLogLevel := logrus.GetLevel()
	defer func() {
		os.Stdout = originalStdout                           // 恢复标准输出
		logrus.StandardLogger().SetOutput(originalLogOutput) // 恢复日志输出
		logrus.SetLevel(originalLogLevel)                    // 恢复日志级别
	}()

	// 2. 重定向标准输出：用管道捕获fmt.Println的内容
	rStdout, wStdout, err := os.Pipe()
	assert.NoError(t, err, "创建标准输出管道失败")
	os.Stdout = wStdout // 将标准输出指向管道写入端

	// 3. 重定向日志输出：用缓冲区捕获logrus.Debug的内容
	var logBuf bytes.Buffer
	logrus.SetOutput(&logBuf)          // 日志写入缓冲区
	logrus.SetLevel(logrus.DebugLevel) // 确保Debug级别日志能输出

	// 4. 模拟执行version命令的Run逻辑
	versionCmd.Run(&cobra.Command{}, []string{}) // 空args（version无需参数）

	// 5. 读取管道数据到缓冲区（关键修复：用io.Copy替代bytes.ReadFrom）
	wStdout.Close() // 先关闭写入端，避免读取阻塞
	var stdoutBuf bytes.Buffer
	_, err = io.Copy(&stdoutBuf, rStdout) // 从rStdout复制数据到stdoutBuf
	assert.NoError(t, err, "读取标准输出管道失败")

	// 6. 断言标准输出（fmt.Println的结果）
	expectedOutput := fmt.Sprintf("%s %s\n", cliName, Version)
	assert.Equal(t, expectedOutput, stdoutBuf.String(), "版本命令的标准输出不符合预期")

	// 7. 断言日志输出（logrus.Debug的结果）
	assert.Contains(t, logBuf.String(), "current version", "日志应包含版本提示文本")
	assert.Contains(t, logBuf.String(), Version, "日志应包含版本号")
	assert.Contains(t, logBuf.String(), "version", "日志应包含version字段")
}
