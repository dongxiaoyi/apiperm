package utils

import (
	"encoding/base64"
	"go.uber.org/zap"
	"golang.org/x/crypto/scrypt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"errors"
)

// 获取当前操作路径的目录
func AbsPath() string {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return ""
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return ""
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		panic(errors.New(`error: Can't find "/" or "\".`))
		return ""
	}
	return string(path[0 : i+1])
}

// 获取当前时间
func CurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 密码加密
func Decrypt(passwd string, logger *zap.SugaredLogger) string {
	salt := []byte("I'm salt")
	dk, err := scrypt.Key([]byte(passwd), salt, 1<<15, 8, 1, 32)
	if err != nil {
		logger.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(dk)
}
