# QuickFileServe-GoLang

一个使用 Go 语言编写的简单文件服务器命令行工具，可以快速启动一个 HTTP 文件服务器来共享文件。

## 功能特点

- 快速启动本地文件服务器
- 支持文件上传/下载
- 自动生成访问二维码（可选）
- 健康检查接口
- 防止目录遍历攻击
- JSON 格式的日志输出

## 安装

```bash
make build
```

编译后的二进制文件将生成在 `build/file-server-cli` 目录下。

## 使用方法

### 基本命令

```bash
# 查看版本
file-server-cli version

# 启动文件服务器
file-server-cli serve [flags]
```

### 服务器命令选项

```bash
# 在指定端口启动服务器并分享当前目录
file-server-cli serve -p 8080 -d .

# 启动服务器并生成二维码
file-server-cli serve -d /path/to/share -q true

# 生成小型二维码
file-server-cli serve -d . -q small
```

参数说明：
- `-p, --port`: 指定监听端口（默认：9999）
- `-d, --dir`: 指定要共享的目录（默认：当前目录）
- `-q, --qrcode`: 是否生成二维码 (可选值: 'false', 'true', 'small')
- `--log-level`: 设置日志级别 (debug, info, warn, error, fatal, panic)

## API 接口

- `GET /files/*`: 访问共享文件
- `GET /list`: 获取文件列表
- `POST /upload`: 上传文件
- `GET /ping`: 存活检测
- `GET /health`: 健康检查

## 开发

清理构建:
```bash
make clean
```

## 注意事项

- 默认监听所有网络接口 (0.0.0.0)
- 文件上传大小限制为 100MB
- 支持通过 Ctrl+C 优雅关闭服务

## 版本

当前版本:`v0.0.1`