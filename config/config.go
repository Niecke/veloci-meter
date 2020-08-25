package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	l "github.com/sirupsen/logrus"
)

type Config struct {
	MailURI       string `json:"MailURI"`
	MailUser      string `json:"MailUser"`
	MailPassword  string `json:"MailPassword"`
	BatchSize     int    `json:"BatchSize"`
	FetchInterval int    `json:"FetchInterval"`
	CheckInterval int    `json:"CheckInterval"`
	LogLevel      string `json:"LogLevel"`

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

func LoadConfig() (c *Config) {
	//##### CONFIG #####
	var config Config
	configFile, err := os.Open("/opt/veloci-meter/config.json")
	if err != nil {
		l.Fatal(err)
	}
	l.Debugln("Successfully Opened config.json")
	defer configFile.Close()

	byteValue, _ := ioutil.ReadAll(configFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'config' which we defined above
	json.Unmarshal(byteValue, &config)
	l.Infoln("Successfully loaded the config from /opt/veloci-meter/config.json")
	return &config
}
