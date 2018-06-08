package main

import (
	"os"

	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("hera")

// InitLogger configures logging
func InitLogger() {
	logFile, err := os.OpenFile("/var/log/hera/hera.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Errorf("Unable open log file: %s", err)
	}

	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	strderrBackendFormat := logging.MustStringFormatter(`[%{level}] %{message}`)
	strderrBackendFormatter := logging.NewBackendFormatter(stderrBackend, strderrBackendFormat)

	logFileBackend := logging.NewLogBackend(logFile, "", 0)
	logFileBackendFormat := logging.MustStringFormatter(`%{time:15:04:00.000} [%{level}] %{message}`)
	logFileBackendFormatter := logging.NewBackendFormatter(logFileBackend, logFileBackendFormat)

	logging.SetBackend(strderrBackendFormatter, logFileBackendFormatter)
}
