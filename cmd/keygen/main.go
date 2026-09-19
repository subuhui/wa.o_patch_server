package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"shorebird-server/internal/crypto"
)

func main() {
	outDir := flag.String("out", ".", "Output directory for generated keys")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	privKey, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	pemBytes := crypto.ExportPrivateKeyPEM(privKey)
	privKeyPath := filepath.Join(*outDir, "private_key.pem")
	if err := os.WriteFile(privKeyPath, pemBytes, 0600); err != nil {
		log.Fatalf("Failed to write private key to %s: %v", privKeyPath, err)
	}

	pubKeyBase64, err := crypto.ExportPublicKeyDERBase64(&privKey.PublicKey)
	if err != nil {
		log.Fatalf("Failed to export public key: %v", err)
	}

	pubKeyPath := filepath.Join(*outDir, "public_key.der.base64")
	if err := os.WriteFile(pubKeyPath, []byte(pubKeyBase64), 0644); err != nil {
		log.Fatalf("Failed to write public key to %s: %v", pubKeyPath, err)
	}

	fmt.Println("==================================================")
	fmt.Printf("✅ RSA 2048 密钥对生成成功！\n")
	fmt.Printf("私钥已保存至: %s\n", privKeyPath)
	fmt.Printf("公钥已保存至: %s\n", pubKeyPath)
	fmt.Println("--------------------------------------------------")
	fmt.Println("请在客户端 Flutter 工程的 shorebird.yaml 中添加:")
	fmt.Printf("patch_public_key: %s\n", pubKeyBase64)
	fmt.Println("==================================================")
}
