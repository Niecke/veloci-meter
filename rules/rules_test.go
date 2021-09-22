package rules

import (
	"testing"

	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/test"
)

func TestToString(t *testing.T) {
	s := LoadRules("rules.example.json").Rules[0].ToString()
	test.CheckResult(t, s, "Name: 'full' | Pattern: 'full' | Timeframe: '10' | Ok: '0' | Warning: '1' | Critical: '5'")
}

func TestLoadRulesGlobal5m(t *testing.T) {
	r := LoadRules("rules.example.json")
	test.CheckResult(t, r.Global.FiveMinutes, 10)
}

func TestLoadRulesGlobal60m(t *testing.T) {
	r := LoadRules("rules.example.json")
	test.CheckResult(t, r.Global.SixtyMinutes, 50)
}

func TestLoadRulesFull(t *testing.T) {
	r := LoadRules("rules.example.json").Rules[0]
	test.CheckResult(t, r.Name, "full")
	test.CheckResult(t, r.Pattern, "full")
	test.CheckResult(t, r.Timeframe, 10)
	test.CheckResult(t, r.Warning, int64(1))
	test.CheckResult(t, r.Critical, int64(5))
	test.CheckResult(t, r.Ok, int64(0))
}

func TestLoadRulesWarning(t *testing.T) {
	r := LoadRules("rules.example.json").Rules[1]
	test.CheckResult(t, r.Name, "warning")
	test.CheckResult(t, r.Pattern, "warning")
	test.CheckResult(t, r.Timeframe, 15)
	test.CheckResult(t, r.Warning, int64(3))
	test.CheckResult(t, r.Critical, int64(0))
	test.CheckResult(t, r.Ok, int64(0))
}

func TestLoadRulesCritical(t *testing.T) {
	r := LoadRules("rules.example.json").Rules[2]
	test.CheckResult(t, r.Name, "critical")
	test.CheckResult(t, r.Pattern, "critical")
	test.CheckResult(t, r.Timeframe, 20)
	test.CheckResult(t, r.Warning, int64(0))
	test.CheckResult(t, r.Critical, int64(5))
	test.CheckResult(t, r.Ok, int64(0))
}

func TestLoadRulesOk(t *testing.T) {
	r := LoadRules("rules.example.json").Rules[3]
	test.CheckResult(t, r.Name, "ok")
	test.CheckResult(t, r.Pattern, "ok")
	test.CheckResult(t, r.Timeframe, 3600)
	test.CheckResult(t, r.Warning, int64(0))
	test.CheckResult(t, r.Critical, int64(0))
	test.CheckResult(t, r.Ok, int64(1))
}

func TestLoadRulesWarningError(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadRules("rules.warning.error.json")
	test.CheckResult(t, fatal, true)
}

func TestLoadRulesCriticalError(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadRules("rules.critical.error.json")
	test.CheckResult(t, fatal, true)
}

func TestLoadRulesMissing(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadRules("rules.missing.json")
	test.CheckResult(t, fatal, true)
}

func TestLoadRulesMissingTimeframe(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadRules("rules.missing-timeframe.json")
	test.CheckResult(t, fatal, true)
}

func TestLoadRulesIOError(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadRules("rules")
	test.CheckResult(t, fatal, true)
}

func TestLoadRulesParseError(t *testing.T) {
	defer func() { l.StandardLogger().ExitFunc = nil }()
	var fatal bool
	l.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	LoadRules("rules.broken.json")
	test.CheckResult(t, fatal, true)
}
