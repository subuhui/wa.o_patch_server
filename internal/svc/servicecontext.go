package svc

import (
	"crypto/rsa"
	"log"
	"os"
	"path/filepath"

	"github.com/zeromicro/go-zero/rest"
	"gorm.io/gorm"

	"shorebird-server/internal/config"
	"shorebird-server/internal/crypto"
	"shorebird-server/internal/db"
	"shorebird-server/internal/middleware"
	"shorebird-server/internal/storage"
)

type ServiceContext struct {
	Config         config.Config
	DB             *gorm.DB
	Storage        storage.Storage
	PrivateKey     *rsa.PrivateKey
	AuthMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 1. Initialize RSA Key
	var privKey *rsa.PrivateKey
	if c.Crypto.PrivateKeyPath != "" {
		if _, err := os.Stat(c.Crypto.PrivateKeyPath); os.IsNotExist(err) {
			log.Printf("⚠️ 私钥 %s 不存在，自动生成 RSA 2048 密钥对...", c.Crypto.PrivateKeyPath)
			_ = os.MkdirAll(filepath.Dir(c.Crypto.PrivateKeyPath), 0755)
			var err error
			privKey, err = crypto.GenerateRSAKeyPair()
			if err != nil {
				log.Fatalf("❌ 自动生成 RSA 密钥失败: %v", err)
			}
			pemData := crypto.ExportPrivateKeyPEM(privKey)
			if err := os.WriteFile(c.Crypto.PrivateKeyPath, pemData, 0600); err != nil {
				log.Fatalf("❌ 写入私钥文件失败: %v", err)
			}
			pubKeyBase64, _ := crypto.ExportPublicKeyDERBase64(&privKey.PublicKey)
			pubKeyFile := filepath.Join(filepath.Dir(c.Crypto.PrivateKeyPath), "public_key.der.base64")
			_ = os.WriteFile(pubKeyFile, []byte(pubKeyBase64), 0644)
			log.Printf("✅ 私钥已保存: %s", c.Crypto.PrivateKeyPath)
			log.Printf("🔑 客户端公钥已导出: %s", pubKeyFile)
		} else {
			var err error
			privKey, err = crypto.LoadPrivateKeyFromFile(c.Crypto.PrivateKeyPath)
			if err != nil {
				log.Fatalf("❌ 加载私钥文件失败: %v", err)
			}
			log.Printf("🔐 私钥加载成功: %s", c.Crypto.PrivateKeyPath)
		}
	}

	// 2. Initialize Database
	database, err := db.InitDB(c.Database.Type, c.Database.DSN)
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	log.Printf("🗄️ 数据库连接成功 (%s)", c.Database.Type)

	// 3. Initialize Storage
	var st storage.Storage
	switch c.Storage.Type {
	case "local":
		st, err = storage.NewLocalStorage("data/storage", c.PublicURL)
		if err != nil {
			log.Fatalf("❌ 本地磁盘存储初始化失败: %v", err)
		}
		log.Println("📦 已启用本地磁盘存储模式 (data/storage)")
	default:
		st, err = storage.NewMinIOStorage(&c)
		if err != nil {
			log.Fatalf("❌ MinIO 对象存储初始化失败: %v", err)
		}
		log.Printf("📦 MinIO 对象存储已连接 (Endpoint: %s, Bucket: %s)", c.Storage.MinIO.Endpoint, c.Storage.MinIO.Bucket)
	}

	return &ServiceContext{
		Config:         c,
		DB:             database,
		Storage:        st,
		PrivateKey:     privKey,
		AuthMiddleware: middleware.NewAuthMiddleware(c.Auth.Token).Handle,
	}
}
