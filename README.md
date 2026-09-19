# Shorebird private server

基于 Go、go-zero、GORM/MySQL 的 Shorebird 私有补丁服务，支持本地磁盘或 MinIO 存储。

## 开发与运行

需要 Go 1.26.5 或更高版本、MySQL；测试使用 SQLite，需要启用 CGO 和 C 编译器。

```sh
go test ./...
go build -o bin/shorebird-server ./cmd/server
go run ./cmd/keygen -out keys
```

在本机创建 `etc/shorebird-api.yaml`，填写以下配置。占位符必须替换为实际值；不应将配置和私钥提交到版本库。

```yaml
Name: shorebird-api
Host: 0.0.0.0
Port: 8080
Timeout: 60000
MaxBytes: 536870912
PublicURL: https://YOUR_SERVER
Auth:
  Token: REPLACE_WITH_PRIVATE_TOKEN
Crypto:
  PrivateKeyPath: keys/private_key.pem
Database:
  Type: mysql
  DSN: "USER:PASSWORD@tcp(MYSQL_HOST:3306)/DATABASE?charset=utf8mb4&parseTime=True&loc=Local"
Storage:
  Type: local
  MinIO: {}
```

```sh
./bin/shorebird-server -f etc/shorebird-api.yaml
```

选择 MinIO 时，将 `Storage.Type` 设为 `minio`，并配置 `Storage.MinIO` 中的 `Endpoint`、`AccessKeyID`、`SecretAccessKey`、`UseSSL`、`Bucket`、`ProxyDownload`。本地存储目录为工作目录下的 `data/storage`。

所有 YAML、私钥、`keys/`、`data/` 和编译产物均已通过 `.gitignore` 排除。

## 补丁发布行为

1. 创建补丁后状态为 `draft`，不会影响正在下发的补丁。
2. 创建各平台、架构的产物记录并完成上传。本地上传采用临时文件与原子重命名，失败时不会暴露半成品。
3. 调用 promote 并提供属于该应用的 `channel_id`。服务端检查已登记的产物全部存在且大小匹配后，才激活补丁。
4. 更新检查按渠道、平台及架构选择最新已发布补丁；未指定渠道时使用 `stable`。已发布补丁的产物元数据不可修改。

升级已有数据库时，自动迁移会增加可空的 `patches.channel_id`。旧数据没有可靠的渠道信息，因此不会自动猜测渠道；需要通过 promote 将旧补丁明确发布到目标渠道后，才会再次向客户端提供。回滚编号仍在 release 范围内返回，以便切换渠道的设备也能撤销已安装的坏补丁。

## 验证

```sh
go test -race ./...
go vet ./...
```

回归测试覆盖目录穿越（含符号链接）、失败上传、生成下载链接的路由匹配、渠道/平台/架构筛选以及草稿发布、文件完整性与回滚流程。逻辑测试使用独立 SQLite 数据库，不连接部署数据库或 MinIO。
