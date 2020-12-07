package config

import (
	"testing"

	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/test"
)

func TestLoadConfigExample(t *testing.T) {
	conf := LoadConfig("config.example.json")

	test.CheckResult(t, *conf.InsecureSkipVerify, true)
	test.CheckResult(t, conf.Mail.URI, "mail.local:993")
	test.CheckResult(t, conf.Mail.User, "test@local")
	test.CheckResult(t, conf.Mail.Password, "xxxxxx")
	test.CheckResult(t, conf.Mail.BatchSize, 5)
	test.CheckResult(t, conf.FetchInterval, 10)
	test.CheckResult(t, conf.CheckInterval, 10)
	test.CheckResult(t, conf.LogLevel, "INFO")
	test.CheckResult(t, conf.LogFormat, "JSON")
	test.CheckResult(t, conf.CleanUpSchedule, "0 * * * *")
	test.CheckResult(t, conf.Icinga.Endpoint, "https://localhost:5665/v1/actions/process-check-result")
	test.CheckResult(t, conf.Icinga.User, "root")
	test.CheckResult(t, conf.Icinga.Password, "xxxxxxx")
	test.CheckResult(t, conf.Icinga.Hostname, "MAIL")
	test.CheckResult(t, conf.Redis.URI, "localhost:6379")
	test.CheckResult(t, conf.Redis.Password, "")
	test.CheckResult(t, conf.Redis.Database, 0)
}

func TestLoadConfigMinimum(t *testing.T) {
	conf := LoadConfig("config.minimum.json")

	test.CheckResult(t, *conf.InsecureSkipVerify, false)
	test.CheckResult(t, conf.Mail.URI, "mail.local:993")
	test.CheckResult(t, conf.Mail.User, "test@local")
	test.CheckResult(t, conf.Mail.Password, "xxxxxxx")
	test.CheckResult(t, conf.Mail.BatchSize, 5)
	test.CheckResult(t, conf.FetchInterval, 10)
	test.CheckResult(t, conf.CheckInterval, 10)
	test.CheckResult(t, conf.LogLevel, "INFO")
	test.CheckResult(t, conf.LogFormat, "PLAIN")
	test.CheckResult(t, conf.CleanUpSchedule, "0 * * * *")
	test.CheckResult(t, conf.Icinga.Endpoint, "https://localhost:5665/v1/actions/process-check-result")
	test.CheckResult(t, conf.Icinga.User, "root")
	test.CheckResult(t, conf.Icinga.Password, "xxxxxxx")
	test.CheckResult(t, conf.Icinga.Hostname, "MAIL")
	test.CheckResult(t, conf.Redis.URI, "localhost:6379")
	test.CheckResult(t, conf.Redis.Password, "")
	test.CheckResult(t, conf.Redis.Database, 0)
}

func TestLoadConfigBrokent(t *testing.T) {
	conf := LoadConfig("config.broken.json")

	test.CheckResult(t, conf.Mail.URI, "mail.local:993")
	test.CheckResult(t, conf.Mail.User, "test@local")
	test.CheckResult(t, conf.Mail.Password, "xxxxxx")
	test.CheckResult(t, conf.Mail.BatchSize, 5)
	test.CheckResult(t, conf.FetchInterval, 10)
	test.CheckResult(t, conf.CheckInterval, 10)
	test.CheckResult(t, conf.LogLevel, "INFO")
	test.CheckResult(t, conf.LogFormat, "PLAIN")
	test.CheckResult(t, conf.CleanUpSchedule, "0 * * * *")
	test.CheckResult(t, conf.Icinga.Endpoint, "https://localhost:5665/v1/actions/process-check-result")
	test.CheckResult(t, conf.Icinga.User, "root")
	test.CheckResult(t, conf.Icinga.Password, "xxxxxxx")
	test.CheckResult(t, conf.Icinga.Hostname, "MAIL")
	test.CheckResult(t, conf.Redis.URI, "localhost:6379")
	test.CheckResult(t, conf.Redis.Password, "")
	test.CheckResult(t, conf.Redis.Database, 0)
}

func TestLoadConfigSyntax(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadConfig("config.syntax.json")
	test.CheckResult(t, fatal, true)
}

func TestLoadConfigMissing(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadConfig("config.missing.json")
	test.CheckResult(t, fatal, true)
}

func TestLoadConfigIOError(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadConfig("config")
	test.CheckResult(t, fatal, true)
}
