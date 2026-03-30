package logger

import "log"

func Info(message string, args ...any) {
	log.Printf("[INFO] "+message, args...)
}

func Error(message string, args ...any) {
	log.Printf("[ERROR] "+message, args...)
}
