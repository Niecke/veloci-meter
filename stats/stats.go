package stats

import (
	"encoding/json"
	"math"
	"os"
	"time"

	"niecke-it.de/veloci-meter/config"
	l "niecke-it.de/veloci-meter/logging"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
)

func ExportJob(conf *config.Config, rules rules.Rules) {
	// get a timestamp from yesterday
	timestamp := int(time.Now().Unix()) - 24*60*60
	timestampDay := int64(timestamp - int(math.Mod(float64(timestamp), float64(24*60*60))))
	m := map[string]interface{}{"time": time.Unix(timestampDay, 0).Format(time.RFC3339), "stats": nil}
	statsList := []rdb.Stats{}
	r := rdb.NewClient(&conf.Redis)

	for _, rule := range rules.Rules {
		stats := r.GetStatisticCount(rule.Name, timestamp)
		statsList = append(statsList, stats)
		l.DebugLog("Appended Stats to list: {{.stats_list}}", map[string]interface{}{
			"stats_list": statsList,
			"stats":      stats})
	}
	m["stats"] = statsList

	b, err := json.Marshal(m)
	if err != nil {
		l.ErrorLog(err, "Error while marshaling stats to JSON", map[string]interface{}{
			"stats_list": statsList})
	} else {
		l.DebugLog("Stats marshaled.", map[string]interface{}{
			"stats_list":            statsList,
			"string_representation": b})
		writeToFile("stats", string(b))
	}
	l.InfoLog("Stats reported to file.", map[string]interface{}{
		"stats_day": time.Unix(timestampDay, 0).Format(time.RFC3339)})
}

func writeToFile(path string, content string) {
	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.ErrorLog(err, "Error while opening stats file at {{.fullpath}}", map[string]interface{}{
			"path": path})
	}
	defer f.Close()
	if n, err := f.WriteString(content + "\n"); err != nil {
		l.ErrorLog(err, "Error while writing stats to file at {{.fullpath}}", map[string]interface{}{
			"path": path})
	} else {
		l.DebugLog("Data written to file.", map[string]interface{}{
			"fullpath":      path,
			"bytes_written": n})
	}
}
