package mocks

type Logger struct{}

func (l *Logger) Debug(msg string, args ...any) {}
func (l *Logger) Warn(msg string, args ...any)  {}
func (l *Logger) Info(msg string, args ...any)  {}
func (l *Logger) Error(msg string, args ...any) {}
