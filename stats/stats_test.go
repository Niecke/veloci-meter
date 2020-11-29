package stats

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"testing"
	"time"

	"niecke-it.de/veloci-meter/config"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
)

func TestExportJobMail(t *testing.T) {
	config := config.LoadConfig("../config.json")
	rs := rules.LoadRules("../rules.example.json")
	r := rdb.NewClient(&config.Redis)

	ts := int(time.Now().Unix()) - (24 * 60 * 60)
	timestampDay := ts - int(math.Mod(float64(ts), float64(24*60*60)))

	for _, rule := range rs.Rules {
		_, err := r.Client().HIncrBy("stats:"+rule.Name+":"+fmt.Sprint(timestampDay), "mail", int64(1)).Result()

		if err != nil {
			t.Errorf("TestExportJob() test returned an unexpected result: [%v]", err)
		}
	}
	ExportJob(config, *rs)

	content, err := ioutil.ReadFile("stats")
	if err != nil {
		t.Errorf("TestExportJob() test returned an unexpected result: [%v]", err)
	}

	result := string(content)
	expected := fmt.Sprintf("{\"stats\":[{\"Name\":\"first rule\",\"Mail\":1,\"Warning\":0,\"Critical\":0},{\"Name\":\"another rule\",\"Mail\":1,\"Warning\":0,\"Critical\":0},{\"Name\":\"positive rule\",\"Mail\":1,\"Warning\":0,\"Critical\":0}],\"time\":\"%v\"}\n", time.Unix(int64(timestampDay), 0).Format(time.RFC3339))
	if result != expected {
		t.Errorf("TestExportJob() test returned an unexpected result: got %v want %v", result, expected)
	}

	os.Remove("stats")
	r.Client().FlushDB()
}