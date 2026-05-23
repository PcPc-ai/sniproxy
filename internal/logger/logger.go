package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志记录器
type Logger struct {
	level  LogLevel
	format string // "text" 或 "json"
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

var defaultLogger *Logger

// InitLogger 初始化默认日志记录器
func InitLogger(level, format string) {
	defaultLogger = NewLogger(level, format)
}

// NewLogger 创建新的日志记录器
func NewLogger(level, format string) *Logger {
	logLevel := parseLogLevel(level)
	if format != "json" {
		format = "text"
	}

	return &Logger{
		level:  logLevel,
		format: format,
	}
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// DebugWithFields 记录带字段的调试日志
func (l *Logger) DebugWithFields(message string, fields map[string]interface{}) {
	l.logWithFields(DEBUG, message, fields)
}

// InfoWithFields 记录带字段的信息日志
func (l *Logger) InfoWithFields(message string, fields map[string]interface{}) {
	l.logWithFields(INFO, message, fields)
}

// WarnWithFields 记录带字段的警告日志
func (l *Logger) WarnWithFields(message string, fields map[string]interface{}) {
	l.logWithFields(WARN, message, fields)
}

// ErrorWithFields 记录带字段的错误日志
func (l *Logger) ErrorWithFields(message string, fields map[string]interface{}) {
	l.logWithFields(ERROR, message, fields)
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)
	l.output(level, message, nil)
}

// logWithFields 内部带字段的日志方法
func (l *Logger) logWithFields(level LogLevel, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	l.output(level, message, fields)
}

// output 输出日志
func (l *Logger) output(level LogLevel, message string, fields map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if l.format == "json" {
		entry := LogEntry{
			Timestamp: timestamp,
			Level:     level.String(),
			Message:   message,
			Fields:    fields,
		}

		jsonData, err := json.Marshal(entry)
		if err != nil {
			log.Printf("JSON序列化失败: %v", err)
			return
		}

		fmt.Println(string(jsonData))
	} else {
		// 文本格式
		logLine := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)

		if len(fields) > 0 {
			var fieldStrs []string
			for k, v := range fields {
				fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
			}
			logLine += fmt.Sprintf(" | %s", strings.Join(fieldStrs, " "))
		}

		fmt.Println(logLine)
	}
}

// 全局日志函数
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	} else {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	} else {
		log.Printf("[INFO] "+format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	} else {
		log.Printf("[WARN] "+format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	} else {
		log.Printf("[ERROR] "+format, args...)
	}
}

func DebugWithFields(message string, fields map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.DebugWithFields(message, fields)
	} else {
		log.Printf("[DEBUG] %s | %+v", message, fields)
	}
}

func InfoWithFields(message string, fields map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.InfoWithFields(message, fields)
	} else {
		log.Printf("[INFO] %s | %+v", message, fields)
	}
}

func WarnWithFields(message string, fields map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.WarnWithFields(message, fields)
	} else {
		log.Printf("[WARN] %s | %+v", message, fields)
	}
}

func ErrorWithFields(message string, fields map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.ErrorWithFields(message, fields)
	} else {
		log.Printf("[ERROR] %s | %+v", message, fields)
	}
}
