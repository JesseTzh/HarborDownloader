# Harbor Downloader

Windows 轻量 Harbor 私有镜像下载与 Docker TAR 导出工具。

在不安装、不启动 Docker Engine 的情况下，从企业内部 Harbor 私有仓库下载容器镜像，并导出为可被 `docker load` 加载的 `.tar` 文件。

## 功能

- 配置 Harbor 地址，使用用户名/密码或 Robot Account 登录
- 创建下载任务，填写任务名称，在任务中配置多个镜像与平台；每个镜像可指定「保存为」的 docker tag，导出 TAR 后 `docker load` 即为该名称
- 在设置中配置默认下载根目录，每个任务自动使用同名子文件夹；再次下载会先删除该文件夹中已有的 TAR 包
- 任务可选「是否打包」：全部 `.tar` 下载完成后合并为 tar archive，再用 zstd -19 压缩为 `all.tar.zst`
- 触发任务后按顺序一次性下载多个镜像并导出 `.tar`
- 显示真实进度（速度、已下载、ETA、当前 layer）
- 取消任务
- 不依赖 Docker Engine / Docker Desktop / WSL

## 技术框架

桌面应用基于 [Wails v2](https://wails.io)，Go 后端与 Vue 前端打包为单个 Windows 可执行文件，运行时通过系统 WebView2 渲染界面。

| 层级 | 技术 |
| --- | --- |
| 桌面壳 | Wails v2.12、WebView2 |
| 后端 | Go 1.23+ |
| 前端 | Vue 3、TypeScript、Vite |
| 镜像协议 | [go-containerregistry](https://github.com/google/go-containerregistry)（OCI / Docker Registry HTTP API V2） |
| 产物 | Docker 兼容 TAR；可选 zstd 打包（[klauspost/compress](https://github.com/klauspost/compress)） |
| 发布 | GitHub Actions、NSIS 安装包 |

主要目录：

| 路径 | 说明 |
| --- | --- |
| `app/` | Wails 绑定与生命周期 |
| `internal/registry/` | Harbor 认证、清单解析、layer 拉取 |
| `internal/download/` | 任务编排、进度、取消 |
| `internal/export/` | TAR 导出 |
| `internal/config/` | 本机配置读写 |
| `frontend/` | Vue 界面 |

## 下载

GitHub Actions 会自动构建 Windows amd64 可执行文件并发布到 [Releases](../../releases)：

| 文件 | 说明 |
| --- | --- |
| `HarborDownloader-windows-amd64.exe` | 便携版，下载后直接运行 |
| `HarborDownloader-windows-amd64-setup.exe` | NSIS 安装包 |
| `SHA256SUMS.txt` | SHA-256 校验 |

- 推送标签 `v1.0.0`：发布正式版
- 推送到 `main`：更新名为 `nightly` 的预发布（会被覆盖），版本号为 `YYMMdd-nightly.x`（`x` 为构建序号）
- 在 Actions 里手动运行 **Release Windows**，可填写版本号（如 `v1.0.0`）

```bash
git tag v1.0.0
git push origin v1.0.0
```

## 文档

| 文档 | 说明 |
| --- | --- |
| [使用说明](docs/usage.md) | 登录 Harbor、创建任务、导出 TAR、配置与日志位置 |
| [开发](docs/development.md) | 环境依赖、本地运行与单元测试 |
| [构建](docs/build.md) | 本机构建、Windows 安装包与版本注入 |
