package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	PublicURL string `json:",optional"`

	Auth struct {
		Token string `json:",optional"`
	}

	Crypto struct {
		PrivateKeyPath string `json:",optional"`
	}

	Database struct {
		Type string `json:",optional"`
		DSN  string `json:",optional"`
	}

	Storage struct {
		Type  string      `json:",optional"`
		MinIO MinIOConfig `json:"MinIO"`
	}
}

type MinIOConfig struct {
	Endpoint          string `json:",optional"`
	AccessKeyID       string `json:",optional"`
	SecretAccessKey   string `json:",optional"`
	UseSSL            bool   `json:",optional"`
	Bucket            string `json:",optional"`
	ProxyDownload     bool   `json:",optional"`
	PublicDownloadURL string `json:",optional"`
}
