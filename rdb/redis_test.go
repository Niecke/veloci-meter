package rdb

import (
	"testing"

	"niecke-it.de/veloci-meter/config"
)

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
		t.Errorf("GlobalCounter5m() test returned an unexpected result: got %v want %v", result, expected)
	}
}
