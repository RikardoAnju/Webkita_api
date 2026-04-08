package middleware

import (
	"log"
	"os"
	"runtime"
	"time"
)

var errorLogger *log.Logger

func InitLogger() {
	if err := os.MkdirAll("./log", 0755); err != nil {
		log.Println("Cannot create log directory, logging to stdout:", err)
		errorLogger = log.New(os.Stdout, "", log.LstdFlags)
		return
	}

	logFile, err := os.OpenFile("./log/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("Cannot open log file, logging to stdout:", err)
		errorLogger = log.New(os.Stdout, "", log.LstdFlags)
		return
	}

	errorLogger = log.New(logFile, "", log.LstdFlags)
}

func LogError(err error, context string) {
	if err == nil {
		return
	}

	// Capture the file and line number
	_, file, line, _ := runtime.Caller(1)

	errorLogger.Printf("[%s] ERROR: %s | Context: %s | File: %s:%d\n",
		time.Now().Format(time.RFC3339), err.Error(), context, file, line)
}
