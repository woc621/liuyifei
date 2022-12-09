package util

import (
	"os"
)
//判断是否是目录，如果是文件或者文件不存在，返回false,目录返回true
func IsDir(file string) bool {
	fi, err := os.Stat(file)
	if err != nil {
		//fmt.Printf("%s文件不存在file %v", file, err)
		return false
	} else {
		return fi.IsDir()
	}
}
