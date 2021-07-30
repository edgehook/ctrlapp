package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"

	"path/filepath"
	"strings"
)

// Caculate the Md5 Sum of this file.
func CaculateMd5Sum(fileName string) (string, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	md5 := md5.New()
	_, err = io.Copy(md5, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(md5.Sum(nil)), nil
}

func FileIsExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

func GetInstallRootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}

	return strings.Replace(dir, "\\", "/", -1)
}
