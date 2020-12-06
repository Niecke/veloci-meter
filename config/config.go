package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	l "niecke-it.de/veloci-meter/logging"
)

type Config struct {
	FetchInterval      int    `json:"FetchInterval,omitempty"`
	CheckInterval      int    `json:"CheckInterval,omitempty"`
	LogLevel           string `json:"LogLevel,omitempty"`
	LogFormat          string `json:"LogFormat,omitempty"`
	CleanUpSchedule    string `json:"CleanUpSchedule,omitempty"`
	StatsPath          string `json:"StatsPath,omitempty"`
	InsecureSkipVerify *bool  `json:"InsecureSkipVerify,omitempty"`

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
		l.FatalLog(err, "There was an error while opening the config file at '{{.path}}'", map[string]interface{}{"path": path})
	}
	l.DebugLog("Successfully Opened config.json", map[string]interface{}{"path": path})
	defer configFile.Close()

	byteValue, _ := ioutil.ReadAll(configFile)

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		l.FatalLog(err, "Can't parse config file '{{.path}}'", map[string]interface{}{"path": path})
	}

	CheckRequiredFields(&config)

	// check for none required configs
	if config.InsecureSkipVerify == nil {
		f := new(bool)
		*f = false
		l.DebugLog("InsecureSkipVerify not set. Using default: true.", map[string]interface{}{})
		config.InsecureSkipVerify = f
	}

	if config.StatsPath == "" {
		l.WarnLog("StatsPath not set. Using default: /var/log/veloci-meter/stats.", map[string]interface{}{})
		config.StatsPath = "/var/log/veloci-meter/stats"
	}

	l.DebugLog("Testing stats file...", map[string]interface{}{"path": path})
	_, err = os.Stat(config.StatsPath)
	if os.IsNotExist(err) {
		p := filepath.Dir(config.StatsPath)
		err = os.MkdirAll(p, 0644)
		if err != nil {
			l.ErrorLog(err, "Error while opening stats file at {{.path}}", map[string]interface{}{"path": p, "fullpath": path})
		}
		file, err := os.Create(config.StatsPath)
		if err != nil {
			l.ErrorLog(err, "Error while opening stats file at {{.path}}", map[string]interface{}{"fullpath": path})
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(config.StatsPath, currentTime, currentTime)
		if err != nil {
			l.ErrorLog(err, "Error while opening stats file at {{.path}}", map[string]interface{}{"fullpath": path})
		}
	}

	if config.LogLevel == "" {
		l.DebugLog("LogLevel not set. Using default: INFO.", map[string]interface{}{})
		config.LogLevel = "INFO"
	} else if !LogLevels[config.LogLevel] {
		l.WarnLog("Log level '{{.log_level}}' not supported. Falling back to INFO.", map[string]interface{}{"log_level": config.LogLevel})
		config.LogLevel = "INFO"
	}

	if config.LogFormat == "" {
		l.DebugLog("LogFormat not set. Using default: PLAIN.", map[string]interface{}{})
		config.LogFormat = "PLAIN"
	} else if !LogFormats[config.LogFormat] {
		l.WarnLog("Log format {{.log_format}} not supported. Falling back to PLAIN.", map[string]interface{}{"log_format": config.LogFormat})
		config.LogFormat = "PLAIN"
	}

	if config.Mail.BatchSize == 0 {
		l.DebugLog("Mail.BatchSize not set. Using default: 5.", map[string]interface{}{})
		config.Mail.BatchSize = 5
	}

	if config.FetchInterval == 0 {
		l.DebugLog("FetchInterval not set. Using default: 10.", map[string]interface{}{})
		config.FetchInterval = 10
	}

	if config.CheckInterval == 0 {
		l.DebugLog("CheckInterval not set. Using default: 10.", map[string]interface{}{})
		config.CheckInterval = 10
	}

	if config.Icinga.Hostname == "" {
		l.DebugLog("Icinga.Hostname not set. Using default: MAIL.", map[string]interface{}{})
		config.Icinga.Hostname = "MAIL"
	}

	if config.Redis.URI == "" {
		l.DebugLog("Redis.URI not set. Using default: localhost:6379.", map[string]interface{}{})
		config.Redis.URI = "localhost:6379"
	}

	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err = p.Parse(config.CleanUpSchedule)

	if err != nil {
		l.WarnLog("Schedule '{{.job_schedule}}' for CleanUpSchedule not supported. Falling back to '0 * * * *'", map[string]interface{}{"error": err, "job_schedule": config.CleanUpSchedule})
		config.CleanUpSchedule = "0 * * * *"
	}

	l.InfoLog("Successfully loaded the config from {{.fullpath}}", map[string]interface{}{"fullpath": path})
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
		l.FatalLog(nil, "ConfigError: %v is undefined or empty", map[string]interface{}{"key_name": keyName})
	}
}
