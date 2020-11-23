package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/robfig/cron/v3"
	l "github.com/sirupsen/logrus"
)

type Config struct {
	MailURI         string `json:"MailURI"`
	MailUser        string `json:"MailUser"`
	MailPassword    string `json:"MailPassword"`
	BatchSize       int    `json:"BatchSize"`
	FetchInterval   int    `json:"FetchInterval"`
	CheckInterval   int    `json:"CheckInterval"`
	LogLevel        string `json:"LogLevel"`
	LogFormat       string `json:"LogFormat"`
	CleanUpSchedule string `json:"CleanUpSchedule"`

	Icinga Icinga `json:"Icinga"`
	Redis  Redis  `json:"Redis"`
	Mail   Mail   `json:"Mail"`
}

type Icinga struct {
	Endpoint string `json:"Endpoint"`
	User     string `json:"User"`
	Password string `json:"Password"`
	Hostname string `json:"Hostname"`
}

type Redis struct {
	URI      string `json:"URI"`
	Password string `json:"Password"`
	Database int    `json:"Database"`
}

type Mail struct {
	URI       string `json:"URI"`
	User      string `json:"User"`
	Password  string `json:"Password"`
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
		l.Fatal(err)
	}
	l.Debugln("Successfully Opened config.json")
	defer configFile.Close()

	byteValue, _ := ioutil.ReadAll(configFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'config' which we defined above
	json.Unmarshal(byteValue, &config)

	if !LogLevels[config.LogLevel] {
		l.Errorf("Log level %v not supported. Falling back to INFO.", config.LogLevel)
		config.LogLevel = "INFO"
	}

	if !LogFormats[config.LogFormat] {
		l.Errorf("Log format %v not supported. Falling back to PLAIN.", config.LogFormat)
		config.LogFormat = "PLAIN"
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
