package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"log"
	"time"
)

func main() {
	fileName := "./test.log"
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	// 打开文件
	tailFile, err := tail.TailFile(fileName, config)
	if err != nil {
		log.Fatalf("打开log文件失败，err = %v", err.Error())
	}
	// 读取数据
	var (
		msg *tail.Line
		ok  bool
	)
	for {
		msg, ok = <-tailFile.Lines
		if !ok {
			log.Println("tail file close reopen , filename : ", tailFile.Filename)
			time.Sleep(time.Second)
			continue
		}
		fmt.Println("msg : ", msg.Text)
	}
}
