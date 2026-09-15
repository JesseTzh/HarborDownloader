package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

const (
	exportKind    = "HarborDownloader.config"
	exportVersion = 1
)

// keySeed is mixed into the AES-256 key derived at runtime. It is not a
// user-supplied secret; it only keeps exported files from containing
// passwords in plaintext.
var keySeed = []byte{
	0x6b, 0x1e, 0xc4, 0x92, 0x0d, 0x57, 0xa8, 0x3f,
	0xe1, 0x14, 0x88, 0xb6, 0x2c, 0x79, 0xd0, 0x45,
	0x9a, 0x03, 0xf7, 0x5e, 0x21, 0xac, 0x68, 0xdf,
	0x37, 0x84, 0x1b, 0xc9, 0x50, 0x0a, 0x73, 0xe6,
}

type exportEnvelope struct {
	Kind string `json:"kind"`
	V    int    `json:"v"`
	N    string `json:"n"`
	C    string `json:"c"`
}

func fileKey() []byte {
	h := sha256.New()
	h.Write([]byte("HarborDownloader config export v1"))
	for i, b := range keySeed {
		h.Write([]byte{b ^ byte(0x5a+i*13)})
	}
	return h.Sum(nil)
}

func newGCM() (cipher.AEAD, error) {
	block, err := aes.NewCipher(fileKey())
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func EncryptExport(plain []byte) ([]byte, error) {
	gcm, err := newGCM()
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, plain, []byte(exportKind))
	env := exportEnvelope{
		Kind: exportKind,
		V:    exportVersion,
		N:    base64.StdEncoding.EncodeToString(nonce),
		C:    base64.StdEncoding.EncodeToString(ct),
	}
	return json.MarshalIndent(env, "", "  ")
}

func openExport(b []byte) ([]byte, error) {
	var env exportEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		return b, nil
	}
	if env.Kind == "" {
		return b, nil
	}
	if env.Kind != exportKind {
		return b, nil
	}
	if env.V != exportVersion {
		return nil, fmt.Errorf("配置文件版本不受支持")
	}
	if env.N == "" || env.C == "" {
		return nil, fmt.Errorf("无法解密配置文件。请使用本程序导出的文件。")
	}
	nonce, err := base64.StdEncoding.DecodeString(env.N)
	if err != nil {
		return nil, fmt.Errorf("无法解密配置文件。请使用本程序导出的文件。")
	}
	ct, err := base64.StdEncoding.DecodeString(env.C)
	if err != nil {
		return nil, fmt.Errorf("无法解密配置文件。请使用本程序导出的文件。")
	}
	gcm, err := newGCM()
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, nonce, ct, []byte(exportKind))
	if err != nil {
		return nil, fmt.Errorf("无法解密配置文件。请使用本程序导出的文件。")
	}
	return plain, nil
}
