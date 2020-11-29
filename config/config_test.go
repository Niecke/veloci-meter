package config

import (
	"testing"

	l "github.com/sirupsen/logrus"
)

func TestLoadConfigExample(t *testing.T) {
	conf := LoadConfig("config.example.json")

	if (*conf.InsecureSkipVerify == true) != true {
		t.Errorf("TestLoadConfigExample() [conf.InsecureSkipVerify] test returned an unexpected result: got %v want %v", *conf.InsecureSkipVerify, true)
	}
	if (conf.Mail.URI == "mail.local:993") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Mail.URI, "mail.local:993")
	}
	if (conf.Mail.User == "test@local") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Mail.User, "test@local")
	}
	if (conf.Mail.Password == "xxxxxx") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Mail.Password, "xxxxxx")
	}
	if (conf.Mail.BatchSize == 5) != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Mail.BatchSize, 5)
	}
	if (conf.FetchInterval == 10) != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.FetchInterval, 10)
	}
	if (conf.CheckInterval == 10) != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.CheckInterval, 10)
	}
	if (conf.LogLevel == "INFO") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.LogLevel, "INFO")
	}
	if (conf.LogFormat == "JSON") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.LogFormat, "JSON")
	}
	if (conf.CleanUpSchedule == "0 * * * *") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.CleanUpSchedule, "0 * * * *")
	}
	if (conf.Icinga.Endpoint == "https://localhost:5665/v1/actions/process-check-result") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Icinga.Endpoint, "https://localhost:5665/v1/actions/process-check-result")
	}
	if (conf.Icinga.User == "root") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Icinga.User, "root")
	}
	if (conf.Icinga.Password == "xxxxxxx") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Icinga.Password, "xxxxxxx")
	}
	if (conf.Icinga.Hostname == "MAIL") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Icinga.Hostname, "MAIL")
	}
	if (conf.Redis.URI == "localhost:6379") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Redis.URI, "localhost:6379")
	}
	if (conf.Redis.Password == "") != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Redis.Password, "")
	}
	if (conf.Redis.Database == 0) != true {
		t.Errorf("TestLoadConfigExample() [] test returned an unexpected result: got %v want %v", conf.Redis.Database, 0)
	}
}

func TestLoadConfigMinimum(t *testing.T) {
	conf := LoadConfig("config.minimum.json")

	if (*conf.InsecureSkipVerify == false) != true {
		t.Errorf("TestLoadConfigExample() [conf.InsecureSkipVerify] test returned an unexpected result: got %v want %v", *conf.InsecureSkipVerify, false)
	}
	if (conf.Mail.URI == "mail.local:993") != true {
		t.Errorf("TestLoadConfigExample() [conf.Mail.URI] test returned an unexpected result: got %v want %v", conf.Mail.URI, "mail.local:993")
	}
	if (conf.Mail.User == "test@local") != true {
		t.Errorf("TestLoadConfigExample() [conf.Mail.User] test returned an unexpected result: got %v want %v", conf.Mail.User, "test@local")
	}
	if (conf.Mail.Password == "xxxxxxx") != true {
		t.Errorf("TestLoadConfigExample() [conf.Mail.Password] test returned an unexpected result: got %v want %v", conf.Mail.Password, "xxxxxx")
	}
	if (conf.Mail.BatchSize == 5) != true {
		t.Errorf("TestLoadConfigExample() [conf.Mail.BatchSize] test returned an unexpected result: got %v want %v", conf.Mail.BatchSize, 5)
	}
	if (conf.FetchInterval == 10) != true {
		t.Errorf("TestLoadConfigExample() [conf.FetchInterval] test returned an unexpected result: got %v want %v", conf.FetchInterval, 10)
	}
	if (conf.CheckInterval == 10) != true {
		t.Errorf("TestLoadConfigExample() [conf.CheckInterval] test returned an unexpected result: got %v want %v", conf.CheckInterval, 10)
	}
	if (conf.LogLevel == "INFO") != true {
		t.Errorf("TestLoadConfigExample() [conf.LogLevel] test returned an unexpected result: got %v want %v", conf.LogLevel, "INFO")
	}
	if (conf.LogFormat == "PLAIN") != true {
		t.Errorf("TestLoadConfigExample() [conf.LogFormat] test returned an unexpected result: got %v want %v", conf.LogFormat, "PLAIN")
	}
	if (conf.CleanUpSchedule == "0 * * * *") != true {
		t.Errorf("TestLoadConfigExample() [conf.CleanUpSchedule] test returned an unexpected result: got %v want %v", conf.CleanUpSchedule, "0 * * * *")
	}
	if (conf.Icinga.Endpoint == "https://localhost:5665/v1/actions/process-check-result") != true {
		t.Errorf("TestLoadConfigExample() [conf.Icinga.Endpoint] test returned an unexpected result: got %v want %v", conf.Icinga.Endpoint, "https://localhost:5665/v1/actions/process-check-result")
	}
	if (conf.Icinga.User == "root") != true {
		t.Errorf("TestLoadConfigExample() [conf.Icinga.User] test returned an unexpected result: got %v want %v", conf.Icinga.User, "root")
	}
	if (conf.Icinga.Password == "xxxxxxx") != true {
		t.Errorf("TestLoadConfigExample() [conf.Icinga.Password] test returned an unexpected result: got %v want %v", conf.Icinga.Password, "xxxxxxx")
	}
	if (conf.Icinga.Hostname == "MAIL") != true {
		t.Errorf("TestLoadConfigExample() [conf.Icinga.Hostname] test returned an unexpected result: got %v want %v", conf.Icinga.Hostname, "MAIL")
	}
	if (conf.Redis.URI == "localhost:6379") != true {
		t.Errorf("TestLoadConfigExample() [conf.Redis.URI] test returned an unexpected result: got %v want %v", conf.Redis.URI, "localhost:6379")
	}
	if (conf.Redis.Password == "") != true {
		t.Errorf("TestLoadConfigExample() [conf.Redis.Password] test returned an unexpected result: got %v want %v", conf.Redis.Password, "")
	}
	if (conf.Redis.Database == 0) != true {
		t.Errorf("TestLoadConfigExample() [conf.Redis.Database] test returned an unexpected result: got %v want %v", conf.Redis.Database, 0)
	}
}

func TestLoadConfigBrokent(t *testing.T) {
	conf := LoadConfig("config.broken.json")

	if (conf.Mail.URI == "mail.local:993") != true {
		t.Errorf("TestLoadConfigExample() [conf.Mail.URI] test returned an unexpected result: got %v want %v", conf.Mail.URI, "mail.local:993")
	}
	if (conf.Mail.User == "test@local") != true {
		t.Errorf("TestLoadConfigExample() [conf.Mail.User] test returned an unexpected result: got %v want %v", conf.Mail.User, "test@local")
	}
	if (conf.Mail.Password == "xxxxxx") != true {
		t.Errorf("TestLoadConfigExample() [conf.Mail.Password] test returned an unexpected result: got %v want %v", conf.Mail.Password, "xxxxxx")
	}
	if (conf.Mail.BatchSize == 5) != true {
		t.Errorf("TestLoadConfigExample() [conf.Mail.BatchSize] test returned an unexpected result: got %v want %v", conf.Mail.BatchSize, 5)
	}
	if (conf.FetchInterval == 10) != true {
		t.Errorf("TestLoadConfigExample() [conf.FetchInterval] test returned an unexpected result: got %v want %v", conf.FetchInterval, 10)
	}
	if (conf.CheckInterval == 10) != true {
		t.Errorf("TestLoadConfigExample() [conf.CheckInterval] test returned an unexpected result: got %v want %v", conf.CheckInterval, 10)
	}
	if (conf.LogLevel == "INFO") != true {
		t.Errorf("TestLoadConfigExample() [conf.LogLevel] test returned an unexpected result: got %v want %v", conf.LogLevel, "INFO")
	}
	if (conf.LogFormat == "PLAIN") != true {
		t.Errorf("TestLoadConfigExample() [conf.LogFormat] test returned an unexpected result: got %v want %v", conf.LogFormat, "PLAIN")
	}
	if (conf.CleanUpSchedule == "0 * * * *") != true {
		t.Errorf("TestLoadConfigExample() [conf.CleanUpSchedule] test returned an unexpected result: got %v want %v", conf.CleanUpSchedule, "0 * * * *")
	}
	if (conf.Icinga.Endpoint == "https://localhost:5665/v1/actions/process-check-result") != true {
		t.Errorf("TestLoadConfigExample() [conf.Icinga.Endpoint] test returned an unexpected result: got %v want %v", conf.Icinga.Endpoint, "https://localhost:5665/v1/actions/process-check-result")
	}
	if (conf.Icinga.User == "root") != true {
		t.Errorf("TestLoadConfigExample() [conf.Icinga.User] test returned an unexpected result: got %v want %v", conf.Icinga.User, "root")
	}
	if (conf.Icinga.Password == "xxxxxxx") != true {
		t.Errorf("TestLoadConfigExample() [conf.Icinga.Password] test returned an unexpected result: got %v want %v", conf.Icinga.Password, "xxxxxxx")
	}
	if (conf.Icinga.Hostname == "MAIL") != true {
		t.Errorf("TestLoadConfigExample() [conf.Icinga.Hostname] test returned an unexpected result: got %v want %v", conf.Icinga.Hostname, "MAIL")
	}
	if (conf.Redis.URI == "localhost:6379") != true {
		t.Errorf("TestLoadConfigExample() [conf.Redis.URI] test returned an unexpected result: got %v want %v", conf.Redis.URI, "localhost:6379")
	}
	if (conf.Redis.Password == "") != true {
		t.Errorf("TestLoadConfigExample() [conf.Redis.Password] test returned an unexpected result: got %v want %v", conf.Redis.Password, "")
	}
	if (conf.Redis.Database == 0) != true {
		t.Errorf("TestLoadConfigExample() [conf.Redis.Database] test returned an unexpected result: got %v want %v", conf.Redis.Database, 0)
	}
}

func TestLoadConfigSyntax(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadConfig("config.syntax.json")
	if fatal != true {
		t.Errorf("TestLoadConfigSyntax() test returned an unexpected result: got %v want %v", fatal, true)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadConfig("config.missing.json")
	if fatal != true {
		t.Errorf("TestLoadConfigSyntax() test returned an unexpected result: got %v want %v", fatal, true)
	}
}

func TestLoadConfigIOError(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadConfig("config")
	if fatal != true {
		t.Errorf("TestLoadConfigSyntax() test returned an unexpected result: got %v want %v", fatal, true)
	}
}
