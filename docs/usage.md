# 使用说明

1. 在「设置」页填写 Harbor 地址、用户名、密码（或 Robot Account），以及是否校验 TLS
2. 在「设置」页选择默认下载根目录并保存
3. 保存配置或测试连接
4. 在「任务」页新建任务，填写任务名称，按需勾选「是否打包」，并添加一个或多个镜像（可分别选择 `linux/amd64` 或 `linux/arm64`）。每个镜像可填写「保存为」，例如将 `project/app-gateway:1.0.0-build123` 保存为 `app-gateway:1.0.0`
5. 保存任务后点击「下载」。任务会在默认根目录下创建同名子文件夹；若该文件夹里已有 TAR 包会先删除再下载。填写了「保存为」时，TAR 文件名与 `docker load` 后的镜像 tag 都会使用该名称
6. 勾选打包时，全部 TAR 下载完成后会合并并压缩为 `all.tar.zst`（zstd -19）
7. 在目标机器对每个产物执行 `docker load -i xxx.tar`；若使用打包文件，先解压 `all.tar.zst` 再加载其中的 TAR。无需再手动 `docker tag`

仓库地址、用户名、密码、TLS 设置和下载任务会写入本机 `config.json`（权限 `0600`）。密码保存后界面不再回显，重新输入才会覆盖。可在「设置」页导入或导出配置：导出文件由程序内置密钥加密，密码不会以明文出现；导入时自动解密。旧版未加密的 JSON 仍可导入。日志会脱敏，不会记录密码。

配置文件：`%APPDATA%\HarborDownloader\config.json`  
日志：`%LOCALAPPDATA%\HarborDownloader\logs\`
