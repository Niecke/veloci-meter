package logging

import (
	"bytes"
	"errors"
	"html/template"
	"os"

	"github.com/sirupsen/logrus"
)

// SetUpLogger initializes the logrus framework and set the output to the provided path and uses information from the config.
func SetUpLogger(path string, logLevel string, logFormat string) {
	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ErrorLog(err, "Error while checking log file.", map[string]interface{}{"path": path})
	}
	//defer f.Close()

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		ErrorLog(err, "Error while parsing LogLevel.", map[string]interface{}{"path": path})
	}
	logrus.SetLevel(level)
	logrus.SetOutput(f)
	if logFormat == "JSON" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	if err != nil {
		DebugLog("Logger setup.", map[string]interface{}{"log_level": logrus.GetLevel(), "log_format": logFormat, "log_output": path})
	}
}

// InfoLog uses logrus to write infos logs with the provided fields.
func InfoLog(msg string, fields map[string]interface{}) {
	if fields == nil {
		formattedMessage := formatMessage(msg, fields)
		logrus.WithFields(fields).Info(formattedMessage)
	} else {
		logrus.Info(msg)
	}
}

// DebugLog uses logrus to write debug logs with the provided fields.
func DebugLog(msg string, fields map[string]interface{}) {
	if fields == nil {
		formattedMessage := formatMessage(msg, fields)
		logrus.WithFields(fields).Debug(formattedMessage)
	} else {
		logrus.Info(msg)
	}
}

// WarnLog uses logrus to write warn logs with the provided fields.
func WarnLog(msg string, fields map[string]interface{}) {
	if fields == nil {
		formattedMessage := formatMessage(msg, fields)
		logrus.WithFields(fields).Warn(formattedMessage)
	} else {
		logrus.Info(msg)
	}
}

// ErrorLog uses logrus to write error logs with the provided fields.
func ErrorLog(err error, msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	if err == nil {
		err = errors.New("Unkown")
	}
	fields["error"] = err
	formattedMessage := formatMessage(msg, fields)
	logrus.WithFields(fields).Error(formattedMessage)
}

// FatalLog uses logrus to write fatal logs with the provided fields.
func FatalLog(err error, msg string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	if err == nil {
		err = errors.New("Unkown")
	}
	fields["error"] = err
	formattedMessage := formatMessage(msg, fields)
	logrus.WithFields(fields).Fatalf(formattedMessage)
}

func formatMessage(msg string, fields map[string]interface{}) string {
	tmpl, err := template.New("").Parse(msg)
	if err != nil {
		logrus.WithFields(fields).Errorf("[%v] %v", err, msg)
	}
	var formattedMessage bytes.Buffer
	if err = tmpl.Execute(&formattedMessage, fields); err != nil {
		logrus.WithFields(fields).Errorf("[%v] %v", err, msg)
	}
	return formattedMessage.String()
}
