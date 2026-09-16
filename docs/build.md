# 构建

本机开发构建：

```bash
wails build
```

Windows 安装包（在 Windows 上，或具备 NSIS 的交叉环境）：

```bash
wails build -platform windows/amd64 -nsis -webview2 download
```

产物：

- `HarborDownloader.exe`
- `HarborDownloader-amd64-installer.exe`（NSIS）

WebView2 缺失时采用 `download` 策略，安装包保持较小体积。

版本信息可通过 ldflags 注入：

```bash
go build -ldflags "-X main.Version=$(date -u +%y%m%d)-nightly.0 -X main.Commit=$(git rev-parse --short HEAD) -X main.BuildAt=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

Wails 构建时同样可在 `-ldflags` 中传入。

也可使用 Task：

```bash
task build
task build:windows
```
