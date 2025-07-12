package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Logger struct {
	logger   *log.Logger
	level    LogLevel
	logFile  *os.File
}

func NewLogger() *Logger {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
	}
	
	// Create log file with current date
	logFileName := fmt.Sprintf("logs/bot_%s.log", time.Now().Format("2006-01-02"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		logFile = nil
	}
	
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	
	return &Logger{
		logger:  logger,
		level:   INFO,
		logFile: logFile,
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	
	var levelStr string
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	case FATAL:
		levelStr = "FATAL"
	}
	
	message := fmt.Sprintf("[%s] %s", levelStr, fmt.Sprintf(format, args...))
	
	// Log to console
	l.logger.Print(message)
	
	// Log to file if available
	if l.logFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fileMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
		l.logFile.WriteString(fileMessage)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

func (l *Logger) LogPayment(userID int64, amount float64, currency string, method string, status string) {
	l.Info("Payment: User=%d, Amount=%.2f %s, Method=%s, Status=%s", userID, amount, currency, method, status)
}

func (l *Logger) LogSubscription(userID int64, planName string, action string) {
	l.Info("Subscription: User=%d, Plan=%s, Action=%s", userID, planName, action)
}

func (l *Logger) LogError(context string, err error) {
	l.Error("%s: %v", context, err)
}

func (l *Logger) LogUserAction(userID int64, action string, details string) {
	l.Info("User Action: User=%d, Action=%s, Details=%s", userID, action, details)
}

func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}
