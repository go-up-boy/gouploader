package gouploader

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CheckExtName(filename string, ext []string) (string, error) {
	suffix := filepath.Ext(filename)
	if len(ext) != 0 {
		for _, v := range ext{
			if suffix == "." + strings.TrimLeft(v, ".") {
				return v, nil
			}
		}
		return "", errors.New(fmt.Sprintf("Uploading %s is not allowed", suffix))
	}
	return suffix, nil
}

func FileMd5(file *os.File) string {
	hash := md5.New()
	io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil))
}