package rdb

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"niecke-it.de/veloci-meter/config"
	"niecke-it.de/veloci-meter/test"
)

func TestStoreMail(t *testing.T) {
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	msg := imap.Message{}
	envelope := imap.Envelope{Subject: "Test", MessageId: "test"}
	msg.Envelope = &envelope

	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)
	r.StoreMail(&msg, 15)

	result := r.CountMail("Test")
	expected := int64(10)
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestCalculateGlobalKey1(t *testing.T) {
	result := calculateGlobalKey(1606044626, 5)
	expected := "global:5:1606044600"
	test.CheckResult(t, result, expected)
}

func TestCalculateGlobalKey2(t *testing.T) {
	result := calculateGlobalKey(1606044600, 5)
	expected := "global:5:1606044600"
	test.CheckResult(t, result, expected)
}
func TestCalculateGlobalKey3(t *testing.T) {
	result := calculateGlobalKey(1606044626, 60)
	expected := "global:60:1606042800"
	test.CheckResult(t, result, expected)
}

func TestCalculateGlobalKey4(t *testing.T) {
	result := calculateGlobalKey(1606042800, 60)
	expected := "global:60:1606042800"
	test.CheckResult(t, result, expected)
}

func TestGlobalCounter5m(t *testing.T) {
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	result := r.GetGlobalCounter(5)
	expected := 0
	test.CheckResult(t, result, expected)

	r.IncreaseGlobalCounter(5)
	r.IncreaseGlobalCounter(5)
	r.IncreaseGlobalCounter(5)
	r.IncreaseGlobalCounter(5)
	r.IncreaseGlobalCounter(5)

	result = r.GetGlobalCounter(5)
	expected = 5
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestGetGlobalCounterParseError(t *testing.T) {
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	timeframe := 5
	timestamp := int(time.Now().Unix())
	redisKey := calculateGlobalKey(timestamp, timeframe)
	_, err := r.client.Set(redisKey, "t", time.Duration(0)).Result()

	if err != nil {
		t.Errorf("TestGetGlobalCounterParseError() test returned an unexpected result: [%v]", err)
	}

	result := r.GetGlobalCounter(5)
	expected := 0
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestGlobalCounter5mEmpty(t *testing.T) {
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	result := r.GetGlobalCounter(5)
	expected := 0
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestGetKeys(t *testing.T) {
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	result := len(r.GetKeys("global:*"))
	expected := 0
	test.CheckResult(t, result, expected)

	r.IncreaseGlobalCounter(5)

	result = len(r.GetKeys("global:*"))
	expected = 1
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestDeleteKey(t *testing.T) {
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	_, err := r.client.Set("Test-Key", 1, time.Duration(60)*time.Second).Result()

	if err != nil {
		t.Errorf("TestDeleteKey() test produced an error while creating test key. [%v]", err)
	}

	result := r.DeleteKey("Test-Key")
	expected := int64(1)
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestDeleteKeyMissing(t *testing.T) {
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	result := r.DeleteKey("Test-Key")
	expected := int64(0)
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestStatisticCountMail(t *testing.T) {
	ts := int(time.Now().Unix())
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	result := r.GetStatisticCount("TEST-Rule-Name", ts).Mail
	expected := int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Warning
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Critical
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.IncreaseStatisticCountMail("TEST-Rule-Name")
	expected = int64(1)
	test.CheckResult(t, result, expected)

	result = r.IncreaseStatisticCountMail("TEST-Rule-Name")
	expected = int64(2)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Mail
	expected = int64(2)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Warning
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Critical
	expected = int64(0)
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestStatisticCountWarning(t *testing.T) {
	ts := int(time.Now().Unix())
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	result := r.GetStatisticCount("TEST-Rule-Name", ts).Mail
	expected := int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Warning
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Critical
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.IncreaseStatisticCountWarning("TEST-Rule-Name")
	expected = int64(1)
	test.CheckResult(t, result, expected)

	result = r.IncreaseStatisticCountWarning("TEST-Rule-Name")
	expected = int64(2)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Warning
	expected = int64(2)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Mail
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Critical
	expected = int64(0)
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestStatisticCountCritical(t *testing.T) {
	ts := int(time.Now().Unix())
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()

	result := r.GetStatisticCount("TEST-Rule-Name", ts).Mail
	expected := int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Warning
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Critical
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.IncreaseStatisticCountCritical("TEST-Rule-Name")
	expected = int64(1)
	test.CheckResult(t, result, expected)

	result = r.IncreaseStatisticCountCritical("TEST-Rule-Name")
	expected = int64(2)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Critical
	expected = int64(2)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Mail
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TEST-Rule-Name", ts).Warning
	expected = int64(0)
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}

func TestStatisticCountParseError(t *testing.T) {
	ts := int(time.Now().Unix())
	timestampDay := ts - int(math.Mod(float64(ts), float64(24*60*60)))
	config := config.LoadConfig("../config/config.example.json")
	r := NewClient(&config.Redis)
	r.client.FlushDB()
	r.client.HSet("stats:TestHash:"+fmt.Sprint(timestampDay), "mail", "t")
	r.client.HSet("stats:TestHash:"+fmt.Sprint(timestampDay), "warning", "t")
	r.client.HSet("stats:TestHash:"+fmt.Sprint(timestampDay), "critical", "t")

	result := r.GetStatisticCount("TestHash", ts).Mail
	expected := int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TestHash", ts).Warning
	expected = int64(0)
	test.CheckResult(t, result, expected)

	result = r.GetStatisticCount("TestHash", ts).Critical
	expected = int64(0)
	test.CheckResult(t, result, expected)

	r.client.FlushDB()
}
