package gouploader

import (
	"errors"
	"fmt"
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