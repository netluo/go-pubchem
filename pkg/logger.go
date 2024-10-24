// Package pkg coding=utf-8
// @Project : go-pubchem
// @Time    : 2023/12/6 11:24
// @Author  : chengxiang.luo
// @File    : logger.go
// @Software: GoLand
package pkg

import (
	l4g "github.com/netluo/log4go"
	"os"
	"path/filepath"
)

var logFile string

func getFilename() string {
	args := os.Args
	filename := args[0]
	filename = filepath.Base(filename)
	return filename
}

func InitMyLogger(logfile string) l4g.Logger {
	if logfile != "" {
		logFile = logfile
	} else {
		logFile = getFilename() + ".log"
	}
	//logFile = getFilename() + ".log"
	log4g := make(l4g.Logger)
	flw := l4g.NewFileLogWriter(logFile, true, true)
	flw.SetFormat("[%Y-%m-%d %H:%M:%S.%o] [%L] (%A) %I")
	log4g.AddFilter("file", l4g.INFO, flw)
	return log4g
}

var Logger = InitMyLogger("./go-pubchem.log")
