package main

import (
	"MysqlEvent/slave"
	"flag"
	"github.com/siddontang/go/log"
)

func main() {
	var filePath string
	flag.StringVar(&filePath, "config", "config.json", "指定配置文件")
	flag.Parse()
	err := slave.Start(filePath)
	if err != nil {
		log.Infof("Start MysqlEvent err:", err)
	}
}
