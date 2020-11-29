package stats

import (
	"encoding/json"
	"math"
	"os"
	"time"

	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/config"
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
		l.WithFields(l.Fields{
			"stats_list": statsList,
			"stats":      stats,
		}).Debugf("Appended Stats to list: %v", statsList)
	}
	m["stats"] = statsList

	b, err := json.Marshal(m)
	if err != nil {
		l.WithFields(l.Fields{
			"error":      err,
			"stats_list": statsList,
		}).Errorf("[%v] Error while marshaling stats to JSON", err)
	} else {
		l.WithFields(l.Fields{
			"stats_list":            statsList,
			"string_representation": b,
		}).Debugf("Stats marshaled.")
		writeToFile("stats", string(b))
	}
	l.WithFields(l.Fields{
		"timestamp": time.Unix(timestampDay, 0).Format(time.RFC3339),
	}).Infof("Stats reported to file.")
}

func writeToFile(path string, content string) {
	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.WithFields(l.Fields{
			"path":  path,
			"error": err,
		}).Errorf("[%v] Error while opening stats file at %v", err, path)
	}
	defer f.Close()
	if n, err := f.WriteString(content + "\n"); err != nil {
		l.WithFields(l.Fields{
			"path":  path,
			"error": err,
		}).Errorf("[%v] Error while writing stats to file at %v", err, path)
	} else {
		l.WithFields(l.Fields{
			"path":          path,
			"bytes_written": n,
		}).Debugf("Data written to file.")
	}
}
