package main

import (
	"log"
	"log/syslog"

	"github.com/abrander/agento/configuration"
)

var logDebug *log.Logger = nil
var logInfo *log.Logger = nil
var logWarning *log.Logger = nil
var logError *log.Logger = nil
var logInitialized = false

func InitLogging(c *configuration.Configuration) {
	if logInitialized == true {
		return
	}

	logDebug, _ = syslog.NewLogger(syslog.LOG_DEBUG, 0)
	logInfo, _ = syslog.NewLogger(syslog.LOG_INFO, 0)
	logWarning, _ = syslog.NewLogger(syslog.LOG_WARNING, 0)
	logError, _ = syslog.NewLogger(syslog.LOG_ERR, 0)
}

func LogDebug(format string, v ...interface{}) {
	if logDebug == nil {
		return
	}

	logDebug.Printf(format, v...)
}

func LogInfo(format string, v ...interface{}) {
	if logInfo == nil {
		return
	}

	logInfo.Printf(format, v...)
}

func LogWarning(format string, v ...interface{}) {
	if logWarning == nil {
		return
	}

	logWarning.Printf(format, v...)
}

func LogError(format string, v ...interface{}) {
	if logError == nil {
		return
	}

	logError.Printf(format, v...)
}
