package util

import (
	"os"
)

func IsDir(file string) bool {
	fi, err := os.Stat(file)
	if err != nil {
		//fmt.Printf("%s文件不存在file %v", file, err)
		return false
	} else {
		return fi.IsDir()
	}
}
