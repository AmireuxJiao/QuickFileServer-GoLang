package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 初始化测试用的Gin引擎（复用实际路由配置）
func setupTestRouter() *gin.Engine {
	r := gin.Default()
	// 注册需要测试的路由（与实际代码保持一致）
	r.GET("/ping", ping)
	r.GET("/health", healthCheck)
	r.GET("/list", listFiles)
	// 其他需要测试的接口（如/list、/upload等）也可在此注册
	return r
}

func TestPingHandler(t *testing.T) {
	r := setupTestRouter()

	// 创建模拟请求：GET /ping
	req := httptest.NewRequest("GET", "/ping", nil)
	// 创建模拟响应写入器
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// 验证响应结果
	assert.Equal(t, http.StatusOK, w.Code)                                       // 状态码应为200
	assert.Equal(t, "PONG", w.Body.String())                                     // 响应体应为"PONG"
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type")) // 验证Content-Type
}

func TestHealthCheckHandler(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

// 测试路径遍历防护中间件（preventPathTraversal）
func TestPreventPathTraversal(t *testing.T) {
	r := gin.Default()
	rootDir := "/test/root" // 模拟根目录

	// 注册带中间件的测试路由
	r.GET("/files/*filepath", preventPathTraversal(rootDir), func(c *gin.Context) {
		c.String(http.StatusOK, "valid path")
	})

	// 测试用例：正常路径（允许访问）
	t.Run("valid path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/files/subdir/file.txt", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试用例：路径遍历攻击（如../../etc/passwd，应被拒绝）
	t.Run("path traversal attack", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/files/../../etc/passwd", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code) // 应返回403
	})
}

// TestListFiles_ValidDir：测试正常场景（目录存在且有文件）
func TestListFiles_ValidDir(t *testing.T) {
	// 1. 保存原dir值（测试后恢复，避免影响其他测试）
	originalDir := dir
	defer func() {
		dir = originalDir // 测试结束后还原全局变量dir
	}()

	// 2. 创建临时目录（测试专用，自动清理）
	// os.MkdirTemp("", "test-list-*")：在系统临时目录创建以"test-list-"开头的目录
	tempDir, err := os.MkdirTemp("", "test-list-*")
	assert.NoError(t, err, "创建临时目录失败")
	defer os.RemoveAll(tempDir) // 测试结束后删除临时目录（无论成功/失败）

	// 3. 在临时目录创建测试文件/子目录（模拟真实场景）
	testFiles := []string{
		"file1.txt",   // 普通文件
		"doc.pdf",     // 普通文件
		"subdir",      // 子目录
		".hiddenfile", // 隐藏文件（os.ReadDir默认会返回，需确认）
	}
	for _, fname := range testFiles {
		fullPath := tempDir + "/" + fname
		if fname == "subdir" {
			// 创建子目录
			assert.NoError(t, os.Mkdir(fullPath, 0755), "创建子目录失败")
		} else {
			// 创建空文件
			file, err := os.Create(fullPath)
			assert.NoError(t, err, "创建测试文件失败")
			file.Close()
		}
	}

	// 4. 配置/list接口的目录为临时目录（修改全局变量dir）
	dir = tempDir

	// 5. 模拟HTTP GET请求：/list
	r := setupTestRouter()
	req := httptest.NewRequest("GET", "/list", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 6. 断言响应结果
	// 6.1 断言状态码为200
	assert.Equal(t, http.StatusOK, w.Code, "正常场景应返回200状态码")

	// 6.2 断言响应体是JSON格式，且包含预期的文件名
	var responseFiles []string
	err = json.Unmarshal(w.Body.Bytes(), &responseFiles)
	assert.NoError(t, err, "响应体不是合法的JSON格式")

	// 排序后对比（os.ReadDir返回的顺序不固定，排序后避免顺序问题导致测试失败）
	sort.Strings(testFiles)
	sort.Strings(responseFiles)
	assert.Equal(t, testFiles, responseFiles, "返回的文件列表与预期不符")
}
