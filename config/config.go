package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	l "github.com/sirupsen/logrus"
)

type Config struct {
	FetchInterval   int    `json:"FetchInterval,omitempty"`
	CheckInterval   int    `json:"CheckInterval,omitempty"`
	LogLevel        string `json:"LogLevel,omitempty"`
	LogFormat       string `json:"LogFormat,omitempty"`
	CleanUpSchedule string `json:"CleanUpSchedule,omitempty"`
	StatsPath       string `json:"StatsPath,omitempty"`

	Icinga Icinga `json:"Icinga"`
	Redis  Redis  `json:"Redis,omitempty"`
	Mail   Mail   `json:"Mail"`
}

type Icinga struct {
	Endpoint string `json:"Endpoint"`
	User     string `json:"User"`
	Password string `json:"Password"`
	Hostname string `json:"Hostname,omitempty"`
}

type Redis struct {
	URI      string `json:"URI,omitempty"`
	Password string `json:"Password,omitempty"`
	Database int    `json:"Database,omitempty"`
}

type Mail struct {
	URI       string `json:"URI,omitempty"`
	User      string `json:"User,omitempty"`
	Password  string `json:"Password,omitempty"`
	BatchSize int    `json:"BatchSize"`
}

var LogLevels = map[string]bool{
	"FATAL":   true,
	"ERROR":   true,
	"WARNING": true,
	"INFO":    true,
	"DEBUG":   true,
}

var LogFormats = map[string]bool{
	"PLAIN": true,
	"JSON":  true,
}

func LoadConfig(path string) (c *Config) {
	//##### CONFIG #####
	var config Config
	configFile, err := os.Open(path)
	if err != nil {
		l.Fatalf("[%v] There was an error while opening the config file at '%v'", err, path)
	}
	l.Debugln("Successfully Opened config.json")
	defer configFile.Close()

	byteValue, _ := ioutil.ReadAll(configFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'config' which we defined above
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		l.Fatalf("[%v] Can't parse config file %v", err, path)
	}

	CheckRequiredFields(&config)

	// check for none required configs
	if config.StatsPath == "" {
		l.Error("StatsPath not set. Using default: /var/log/veloci-meter/stats.")
		config.StatsPath = "/var/log/veloci-meter/stats"
	}

	l.WithFields(l.Fields{
		"path": path,
	}).Debug("Testing stats file...")
	_, err = os.Stat(config.StatsPath)
	if os.IsNotExist(err) {
		p := filepath.Dir(config.StatsPath)
		err = os.MkdirAll(p, 0644)
		if err != nil {
			l.WithFields(l.Fields{
				"fullpath": path,
				"path":     p,
				"error":    err,
			}).Errorf("[%v] Error while opening stats file at %v", err, path)
		}
		file, err := os.Create(config.StatsPath)
		if err != nil {
			l.WithFields(l.Fields{
				"path":  path,
				"error": err,
			}).Errorf("[%v] Error while opening stats file at %v", err, path)
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(config.StatsPath, currentTime, currentTime)
		if err != nil {
			l.WithFields(l.Fields{
				"path":  path,
				"error": err,
			}).Errorf("[%v] Error while opening stats file at %v", err, path)
		}
	}

	if config.LogLevel == "" {
		l.Error("LogLevel not set. Using default: INFO.")
		config.LogLevel = "INFO"
	} else if !LogLevels[config.LogLevel] {
		l.Errorf("Log level %v not supported. Falling back to INFO.", config.LogLevel)
		config.LogLevel = "INFO"
	}

	if config.LogFormat == "" {
		l.Error("LogFormat not set. Using default: PLAIN.")
		config.LogFormat = "PLAIN"
	} else if !LogFormats[config.LogFormat] {
		l.Errorf("Log format %v not supported. Falling back to PLAIN.", config.LogFormat)
		config.LogFormat = "PLAIN"
	}

	if config.Mail.BatchSize == 0 {
		l.Error("Mail.BatchSize not set. Using default: 5.")
		config.Mail.BatchSize = 5
	}

	if config.FetchInterval == 0 {
		l.Error("FetchInterval not set. Using default: 10.")
		config.FetchInterval = 10
	}

	if config.CheckInterval == 0 {
		l.Error("CheckInterval not set. Using default: 10.")
		config.CheckInterval = 10
	}

	if config.Icinga.Hostname == "" {
		l.Error("Icinga.Hostname not set. Using default: MAIL.")
		config.Icinga.Hostname = "MAIL"
	}

	if config.Redis.URI == "" {
		l.Error("Redis.URI not set. Using default: localhost:6379.")
		config.Redis.URI = "localhost:6379"
	}

	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err = p.Parse(config.CleanUpSchedule)

	if err != nil {
		l.Errorf("[%v] Schedule '%v' for CleanUpSchedule not supported. Falling back to '0 * * * *'", err, config.CleanUpSchedule)
		config.CleanUpSchedule = "0 * * * *"
	}

	l.Infof("Successfully loaded the config from %v", path)
	return &config
}

func CheckRequiredFields(c *Config) {
	CheckRequiredField(c.Mail.URI, "Mail.URI")
	CheckRequiredField(c.Mail.User, "Mail.User")
	CheckRequiredField(c.Mail.Password, "Mail.Password")
	CheckRequiredField(c.Icinga.Endpoint, "Icinga.Endpoint")
	CheckRequiredField(c.Icinga.User, "Icinga.User")
	CheckRequiredField(c.Icinga.Password, "Icinga.Password")
}

func CheckRequiredField(key interface{}, keyName string) {
	if key == nil || key == "" {
		l.Fatalf("ConfigError: %v is undefined or empty", keyName)
	}
}
