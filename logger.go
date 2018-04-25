/*
 * Copyright (c) 2017, [Ribose Inc](https://www.ribose.com).
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"os"
	"fmt"
	"path"
	"sync"
	"time"
	"runtime"
	"strings"
)

// log file path
const (
	KAOHI_LOG_FILE = "kaohi.log"
)

// log level
const (
	INFO = "INFO"

	WARN = "WARN"

	ERR = "ERR"
)

// logger type
const (
	CONSOLE string = "console"
	FILE    string = "file"
)

// logger struct
type kLogger struct {
	Logger
}

type Logger struct {
	PrinterType string
	Location    string
}

type LogInstance struct {
	LogType    string
	Message    string
	LoggerInit Logger
}

type callerInfo struct {
	packageName string
	fileName    string
	funcName    string
	line        int
}

// external logger variables
var consoleLogger, fileLogger kLogger
var mux sync.Mutex
var kLogFile *os.File

// init logger
func InitLogger(dir_path string, level int) error {
	// create log directory and set log file path
	if os.MkdirAll(dir_path, 0755) != nil {
		return ErrCreateLogDir
	}
	log_path := path.Join(dir_path, KAOHI_LOG_FILE)

	// create log file
	kLogFile, err := os.OpenFile(log_path, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0644)
	defer kLogFile.Close()
	if err != nil {
		return ErrCreateLogFile
	}

	// get logger for console and file
	consoleLogger = GetLogger()
	fileLogger = GetLogger(FILE, log_path)

	return nil
}

// get logger
func GetLogger(selector ...string) kLogger {
	if len(selector) == 0 {
		return kLogger{Logger{CONSOLE, ""}}
	}
	return kLogger{Logger{selector[0], selector[1]}}
}

func Print(log LogInstance, packageName string, fileName string, lineNumber int, funcName string, tt time.Time) {
	logString := fmt.Sprintf("[%s] [%s] [%s::%s::%s] [%d] %s\n", log.LogType, tt.Format(time.RFC3339), packageName, fileName, funcName, lineNumber, log.Message)

	switch log.LoggerInit.PrinterType {
	case "console":
		fmt.Printf(logString)
	case "file":
		kLogFile.WriteString(logString)
	}
}

func logPrinter(log LogInstance) {
	info := retrieveCallInfo()
	timer := time.Now()
	logPrint(log, info, timer)
}

func logPrint(log LogInstance, info *callerInfo, time time.Time) {
	Print(log, info.packageName, info.fileName, info.line, info.funcName, time)
	if log.LogType == "CRT" {
		os.Exit(1)
	}
}

func retrieveCallInfo() *callerInfo {
	pc, file, line, _ := runtime.Caller(4)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	return &callerInfo{
		packageName: packageName,
		fileName:    fileName,
		funcName:    funcName,
		line:        line,
	}
}

func (log kLogger) Info(message string) {
	logPrinter(LogInstance{LogType: "INF", Message: message, LoggerInit: log.Logger})
}

func (log kLogger) Warn(message string) {
	logPrinter(LogInstance{LogType: "WRN", Message: message, LoggerInit: log.Logger})
}

func (log kLogger) Error(message string) {
	logPrinter(LogInstance{LogType: "ERR", Message: message, LoggerInit: log.Logger})
}

func (log kLogger) Debug(message string) {
	logPrinter(LogInstance{LogType: "DBG", Message: message, LoggerInit: log.Logger})
}

// print info
func DEBUG_INFO(format string, args ...interface{}) {
	var msg string

	if len(args) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, args)
	}
	consoleLogger.Info(msg)
}

// print warning
func DEBUG_WARN(format string, args ...interface{}) {
	var msg string

	if len(args) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, args)
	}
	consoleLogger.Warn(msg)
}

// print errors
func DEBUG_ERR(format string, args ...interface{}) {
	var msg string

	if len(args) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, args)
	}
	consoleLogger.Error(msg)
}
