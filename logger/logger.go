package logger

import (
	"log"
	"os"
	"strings"
)

func init() {
	debug := os.Getenv("DEBUG")

	positiveList = make(map[string]bool)

	positives := strings.Split(debug, ",")
	for _, positive := range positives {
		if positive == "*" {
			printAll = true
		} else {
			positiveList[positive] = true
		}
	}
}

var (
	positiveList map[string]bool
	printAll     bool
)

func Printf(pkg string, format string, args ...interface{}) {
	_, print := positiveList[pkg]
	if print || printAll {
		log.Printf("\033[35m"+pkg+"\033[0m: "+format+"\n", args...)
	}
}

func Red(pkg string, format string, args ...interface{}) {
	Printf(pkg, "\033[31m"+format+"\033[0m", args...)
}

func Yellow(pkg string, format string, args ...interface{}) {
	Printf(pkg, "\033[33m"+format+"\033[0m", args...)
}

func Green(pkg string, format string, args ...interface{}) {
	Printf(pkg, "\033[32m"+format+"\033[0m", args...)
}

func Error(pkg string, format string, args ...interface{}) {
	log.Printf("\033[31m"+pkg+"\033[0m: "+format+"\n", args...)
}
