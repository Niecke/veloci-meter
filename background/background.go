package background

import (
	"time"

	"niecke-it.de/veloci-meter/config"
	"niecke-it.de/veloci-meter/icinga"
	l "niecke-it.de/veloci-meter/logging"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
)

// normal rules have a warning and a critical, if a rule has an ok value it is special
// in case ok is sset for the rule we only check if there are more mails than defined by ok
// if not an alter level defined by the rule is set

// TODO add info in case the result from icinga is empty; this is beacause of missing check definitions
func CheckForAlerts(config *config.Config, rules *rules.Rules) {
	r := rdb.NewClient(&config.Redis)
	for {
		criticalFired := 0
		warningFired := 0
		okFired := 0
		// iterate over all rules
		for i, rule := range rules.Rules {
			actCount := r.CountMail(rule.Pattern)
			//r.RemoveAllAlert(rule.Pattern)
			if rule.Ok != 0 {
				if actCount < rule.Ok {
					if rule.Alert == "critical" {
						//r.StoreAlert(config.AlertInterval, rule.Pattern, "critical")
						icinga.SendResults(config, rule.Name, rule.Pattern, 2, actCount)
						r.IncreaseStatisticCountCritical(rule.Name)
						criticalFired++
					} else {
						//r.StoreAlert(config.AlertInterval, rule.Pattern, "warning")
						icinga.SendResults(config, rule.Name, rule.Pattern, 1, actCount)
						r.IncreaseStatisticCountWarning(rule.Name)
						warningFired++
					}
				} else {
					//r.StoreAlert(config.AlertInterval, rule.Pattern, "ok")
					icinga.SendResults(config, rule.Name, rule.Pattern, 0, actCount)
					okFired++
				}
			} else {
				// remove all alerts if there are any
				// the alert will be set again in each iteration
				if actCount > rule.Critical {
					//r.StoreAlert(config.AlertInterval, rule.Pattern, "critical")
					icinga.SendResults(config, rule.Name, rule.Pattern, 2, actCount)
					r.IncreaseStatisticCountCritical(rule.Name)
					l.DebugLog("Rule {{.rule_name}} is {{.status}}", map[string]interface{}{
						"rule_name": rule.Name,
						"status":    "CRITICAL",
					})
					criticalFired++
				} else if actCount > rule.Warning {
					//r.StoreAlert(config.AlertInterval, rule.Pattern, "warning")
					icinga.SendResults(config, rule.Name, rule.Pattern, 1, actCount)
					r.IncreaseStatisticCountWarning(rule.Name)
					l.DebugLog("Rule {{.rule_name}} is {{.status}}", map[string]interface{}{
						"rule_name": rule.Name,
						"status":    "WARNING",
					})
					warningFired++
				} else {
					//r.StoreAlert(config.AlertInterval, rule.Pattern, "ok")
					icinga.SendResults(config, rule.Name, rule.Pattern, 0, actCount)
					l.DebugLog("Rule {{.rule_name}} is {{.status}}", map[string]interface{}{
						"rule_name": rule.Name,
						"status":    "OK",
					})
					okFired++
				}
			}
		}

		// check global counter 5 minutes
		c5 := r.GetGlobalCounter(5)
		// counter is above the defined limit => send icinga warning
		if c5 > rules.Global.FiveMinutes {
			icinga.SendResults(config, "Global 5m", "Global 5m", 1, int64(c5))
			r.IncreaseStatisticCountWarning("Global 5m")
			l.DebugLog("Global Rule for {{.timeframe}} is {{.status}}", map[string]interface{}{
				"timeframe": "5m",
				"status":    "WARNING",
			})
		} else {
			icinga.SendResults(config, "Global 5m", "Global 5m", 0, int64(c5))
			l.DebugLog("Global Rule for {{.timeframe}} is {{.status}}", map[string]interface{}{
				"timeframe": "5m",
				"status":    "OK",
			})
		}
		l.DebugLog("Global counter for {{.timeframe}} is at {{.count}}", map[string]interface{}{
			"timeframe": "5m",
			"count":     c5,
			"limit":     rules.Global.FiveMinutes,
		})

		// check global counter 60 minutes
		c60 := r.GetGlobalCounter(60)
		// counter is above the defined limit => send icinga warning
		if c60 > rules.Global.SixtyMinutes {
			icinga.SendResults(config, "Global 60m", "Global 60m", 1, int64(c60))
			r.IncreaseStatisticCountWarning("Global 60m")
			l.DebugLog("Global Rule for {{.timeframe}} is {{.status}}", map[string]interface{}{
				"timeframe": "60m",
				"status":    "WARNING",
			})
		} else {
			icinga.SendResults(config, "Global 60m", "Global 60m", 0, int64(c60))
			l.DebugLog("Global Rule for {{.timeframe}} is {{.status}}", map[string]interface{}{
				"timeframe": "60m",
				"status":    "OK",
			})
		}
		l.DebugLog("Global counter for {{.timeframe}} is at {{.count}}", map[string]interface{}{
			"timeframe": "60m",
			"count":     c60,
			"limit":     rules.Global.SixtyMinutes,
		})

		l.InfoLog("Rule Status. Next run in {{.check_interval}}", map[string]interface{}{
			"OK":             okFired,
			"WARNING":        warningFired,
			"CRITICAL":       criticalFired,
			"check_interval": config.CheckInterval,
		})
		time.Sleep(time.Duration(config.CheckInterval) * time.Second)
	}
}
