package logger

import (
	"os"
	"path/filepath"

	logging "github.com/op/go-logging"
)

const (
	LogDir = "/var/log/hera"
)

func Init(name string) {
	log := logging.MustGetLogger(name)
	logPath := filepath.Join(LogDir, name)

	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	strderrBackendFormat := logging.MustStringFormatter(`[%{level}] %{message}`)
	strderrBackendFormatter := logging.NewBackendFormatter(stderrBackend, strderrBackendFormat)

	logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Errorf("Unable to open file for logging: %s", err)
	}

	logFileBackend := logging.NewLogBackend(logFile, "", 0)
	logFileBackendFormat := logging.MustStringFormatter(`%{time:15:04:00.000} [%{level}] %{message}`)
	logFileBackendFormatter := logging.NewBackendFormatter(logFileBackend, logFileBackendFormat)

	logging.SetBackend(strderrBackendFormatter, logFileBackendFormatter)
}
