package logger

import (
	"log"
)

type Logger struct{ *log.Logger }

func New() *Logger { return &Logger{log.New(log.Writer(), "pubsub ", log.LstdFlags|log.Lshortfile)} }

func (l *Logger) Infow(msg string, kv ...any)  { l.Printf("INFO %s %v", msg, kv) }
func (l *Logger) Errorw(msg string, kv ...any) { l.Printf("ERROR %s %v", msg, kv) }
func (l *Logger) Sync()                        {}
