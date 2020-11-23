package rdb

import (
	"testing"

	"github.com/emersion/go-imap"
	"niecke-it.de/veloci-meter/config"
)

func TestStoreMail(t *testing.T) {
	config := config.LoadConfig("../config.json")
	r := NewRDB(&config.Redis)

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
	if result != expected {
		t.Errorf("TestStoreMail() test returned an unexpected result: got %v want %v", result, expected)
	}

	r.client.FlushDB()
}

func TestCalculateGlobalKey1(t *testing.T) {
	result := calculateGlobalKey(1606044626, 5)
	expected := "global:5:1606044600"
	if result != expected {
		t.Errorf("TestCalculateGlobalKey1() test returned an unexpected result: got %v want %v", result, expected)
	}
}

func TestCalculateGlobalKey2(t *testing.T) {
	result := calculateGlobalKey(1606044600, 5)
	expected := "global:5:1606044600"
	if result != expected {
		t.Errorf("TestCalculateGlobalKey2() test returned an unexpected result: got %v want %v", result, expected)
	}
}
func TestCalculateGlobalKey3(t *testing.T) {
	result := calculateGlobalKey(1606044626, 60)
	expected := "global:60:1606042800"
	if result != expected {
		t.Errorf("TestCalculateGlobalKey3() test returned an unexpected result: got %v want %v", result, expected)
	}
}

func TestCalculateGlobalKey4(t *testing.T) {
	result := calculateGlobalKey(1606042800, 60)
	expected := "global:60:1606042800"
	if result != expected {
		t.Errorf("TestCalculateGlobalKey4() test returned an unexpected result: got %v want %v", result, expected)
	}
}

func TestGlobalCounter5m(t *testing.T) {
	config := config.LoadConfig("../config.json")
	r := NewRDB(&config.Redis)

	r.IncreaseGlobalCounter(5)
	r.IncreaseGlobalCounter(5)
	r.IncreaseGlobalCounter(5)
	r.IncreaseGlobalCounter(5)
	r.IncreaseGlobalCounter(5)

	result := r.GetGlobalCounter(5)
	expected := 5
	if result != expected {
		t.Errorf("TestGlobalCounter5m() test returned an unexpected result: got %v want %v", result, expected)
	}

	r.client.FlushDB()
}

func TestGetKeys(t *testing.T) {
	config := config.LoadConfig("../config.json")
	r := NewRDB(&config.Redis)

	result := len(r.GetKeys("global:*"))
	expected := 0
	if result != expected {
		t.Errorf("TestGetKeys() test returned an unexpected result: got %v want %v", result, expected)
	}

	r.IncreaseGlobalCounter(5)

	result = len(r.GetKeys("global:*"))
	expected = 1
	if result != expected {
		t.Errorf("TestGetKeys() test returned an unexpected result: got %v want %v", result, expected)
	}

	r.client.FlushDB()
}
