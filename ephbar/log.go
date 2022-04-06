package main

import (
	"log"
)

var LogError = 1
var LogInfo = 2
var LogVerbose = 4

var LogLevel = LogError | LogInfo | LogVerbose

func info(format string, v ...interface{}) {
	if LogLevel&LogInfo != 0 {
		log.Printf("\u001b[33m[INFO] "+format+"\u001b[0m\n", v...)
	}
}

func verbose(format string, v ...interface{}) {
	if LogLevel&LogVerbose != 0 {
		log.Printf("\u001b[37m[INFO] "+format+"\u001b[0m\n", v...)
	}
}

func fatal(format string, v ...interface{}) {
	if LogLevel&LogError != 0 {
		log.Fatalf("\u001b[31m[FATAL] "+format+"\u001b[0m\n", v...)
	}
}
