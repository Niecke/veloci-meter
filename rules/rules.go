package rules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	l "github.com/sirupsen/logrus"
)

type Rules struct {
	Global Global `json:"global"`
	Rules  []Rule `json:"rules"`
}

type Global struct {
	FiveMinutes  int `json:"5"`
	SixtyMinutes int `json:"60"`
}

type Rule struct {
	Name      string `json:"name"`
	Pattern   string `json:"pattern"`
	Timeframe int    `json:"timeframe"`
	Warning   int64  `json:"warning"`
	Critical  int64  `json:"critical"`
	Ok        int64  `json:"ok"`
	Alert     string `json:"alert"`
}

type Ok int64
type Alert string

func (r *Rule) ToString() string {
	return "Pattern: '" + r.Pattern + "' | Timeframe: '" + fmt.Sprint(+r.Timeframe) + "' | Ok: '" + fmt.Sprint(r.Ok) + "' | Warning: '" + fmt.Sprint(r.Warning) + "' | Critical: '" + fmt.Sprint(r.Critical) + "'"
}

func LoadRules(path string) (r *Rules) {
	// Open our jsonFile
	jsonFile, err := os.Open(path)
	if err != nil {
		l.Fatal(err)
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our rules array
	var rules Rules

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'rules' which we defined above
	json.Unmarshal(byteValue, &rules)
	l.Infoln("Successfully loaded the rules from %v", path)
	return &rules
}
