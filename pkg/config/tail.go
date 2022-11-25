package config

import (
	"fmt"

	"github.com/hpcloud/tail"
)

func ReadLogByTail(filename string) (chan *tail.Line, error) {
	//filename := "./xx.log"
	fileconfig := tail.Config{
		ReOpen:    true,
		MustExist: false,
		Follow:    true,
		Poll:      true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
	}
	filetail, err := tail.TailFile(filename, fileconfig)
	if err != nil {
		fmt.Println("tailfile err:", err)
		return nil, err
	}
	return filetail.Lines, err
	//var msg *tail.Line
	//var ok bool
}
