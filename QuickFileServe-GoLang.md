# QuickFileServe-GoLang

## runServe函数详解

`filepath.Abs(dir)`标准库函数，用来将传入的字符串组建为当前目录下的绝对路径

运行在 `1-QuickFileServer-GoLang`目录中，运行情况

> msg":"filepaht.Abs(dir) : path = /home/luna/3-develop-live/1-QuickFileServer-GoLang"

运行在 `1-QuickFileServer-GoLang/temp`目录中，运行情况

> "msg":"filepaht.Abs(dir) : path = /home/luna/3-develop-live/1-QuickFileServer-GoLang/temp"

`net.InterfaceAddrs()`net标准库函数，用来获取当前系统内所有可用的IP地址

> "msg":"[127.0.0.1/8 10.255.255.254/32 192.168.3.118/24 ::1/128 fe80::51d1:c072:fd7d:959e/64]"

```go
// 在 Go 语言中，a.(*net.IPNet) 是类型断言（Type Assertion） 的语法，
// 用于检查接口变量 a 中存储的实际值是否为 *net.IPNet 类型，并尝试将其转换为该类型
if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
    ips = append(ips, ipNet.IP.String())
}

```

