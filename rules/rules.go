package rules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	l "niecke-it.de/veloci-meter/logging"
)

// Rules contains a list of rules which can be defined in rules.json and the global rules.
type Rules struct {
	Global Global `json:"global"`
	Rules  []Rule `json:"rules"`
}

// Global contains of a limit for the 5 minute and 60 minute timeframe.
type Global struct {
	FiveMinutes  int `json:"5"`
	SixtyMinutes int `json:"60"`
}

// Rule contains of a Name which should be unique, a Pattern for matching mail subjects, a timeframe which defines the duration one mail will be stored in redis.
// There could be a limit for Ok, Warning and Critical.
type Rule struct {
	Name      string `json:"name"`
	Pattern   string `json:"pattern"`
	Timeframe int    `json:"timeframe"`
	Warning   int64  `json:"warning"`
	Critical  int64  `json:"critical"`
	Ok        int64  `json:"ok"`
	Alert     string `json:"alert"`
}

// ToString formats a rule as string for printing it to console.
func (r *Rule) ToString() string {
	return fmt.Sprintf("Name: '%v' | Pattern: '%v' | Timeframe: '%v' | Ok: '%v' | Warning: '%v' | Critical: '%v'", r.Name, r.Pattern, r.Timeframe, r.Ok, r.Warning, r.Critical)
}

// LoadRules loads all rules from a JSON file stored at path and returns a pointer to the struct where these rules are stored.
func LoadRules(path string) (r *Rules) {
	// Open our jsonFile
	jsonFile, err := os.Open(path)
	if err != nil {
		l.FatalLog(err, "Error while opening rules file at '{{.path}}'", map[string]interface{}{"path": path})
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our rules array
	var rules Rules

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'rules' which we defined above
	if err := json.Unmarshal(byteValue, &rules); err != nil {
		l.ErrorLog(err, "Unknown error while unmarshaling rules.", map[string]interface{}{
			"data": byteValue,
		})
	}
	l.InfoLog("Successfully loaded the rules from {{.path}}", map[string]interface{}{"fullpath": path})
	return &rules
}
