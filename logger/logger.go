package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stderr)
}

// Setup 配置日志级别和文件输出
func Setup(printLevel, logfileLevel string, name string) {
	level, err := logrus.ParseLevel(printLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// 创建日志目录
	logDir := "logs"
	os.MkdirAll(logDir, 0755)

	// 生成日志文件名
	formattedDate := time.Now().Format("20060102")
	logName := formattedDate
	if name != "" {
		logName = name + "_" + formattedDate
	}
	logPath := filepath.Join(logDir, logName+".log")

	// 添加文件输出
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		fileLevel, _ := logrus.ParseLevel(logfileLevel)
		log.AddHook(&fileHook{file, fileLevel})
	}
}

type fileHook struct {
	file  *os.File
	level logrus.Level
}

func (h *fileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *fileHook) Fire(entry *logrus.Entry) error {
	if entry.Level <= h.level {
		line, err := entry.String()
		if err != nil {
			return err
		}
		_, err = h.file.WriteString(line)
		return err
	}
	return nil
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return log
}

// 便捷函数
func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Warningf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

