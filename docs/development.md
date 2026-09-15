# 开发

需要：Go 1.23+、Node.js 18+、[Wails CLI v2.12.0](https://wails.io)

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
wails doctor
wails dev
```

单元测试：

```bash
go test ./...
```

也可使用 Task：

```bash
task dev
task test
```
