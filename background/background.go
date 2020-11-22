package background

import (
	"fmt"
	"time"

	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/config"
	"niecke-it.de/veloci-meter/icinga"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
)

// normal rules have a warning and a critical, if a rule has an ok value it is special
// in case ok is sset for the rule we only check if there are more mails than defined by ok
// if not an alter level defined by the rule is set

// TODO add info in case the result from icinga is empty; this is beacause of missing check definitions
func CheckForAlerts(config *config.Config, rules *rules.Rules) {
	r := rdb.NewRDB(&config.Redis)
	for {
		critical_fired := 0
		warning_fired := 0
		ok_fired := 0
		// iterate over all rules
		for i, rule := range rules.Rules {
			actCount := r.CountMail(rule.Pattern)
			//r.RemoveAllAlert(rule.Pattern)
			if rule.Ok != 0 {
				if actCount < rule.Ok {
					if rule.Alert == "critical" {
						//r.StoreAlert(config.AlertInterval, rule.Pattern, "critical")
						icinga.SendResults(config, rule.Name, rule.Pattern, 2)
						critical_fired += 1
					} else {
						//r.StoreAlert(config.AlertInterval, rule.Pattern, "warning")
						icinga.SendResults(config, rule.Name, rule.Pattern, 1)
						warning_fired += 1
					}
				} else {
					//r.StoreAlert(config.AlertInterval, rule.Pattern, "ok")
					icinga.SendResults(config, rule.Name, rule.Pattern, 0)
					ok_fired += 1
				}
			} else {
				// remove all alerts if there are any
				// the alert will be set again in each iteration
				if actCount > rule.Critical {
					//r.StoreAlert(config.AlertInterval, rule.Pattern, "critical")
					icinga.SendResults(config, rule.Name, rule.Pattern, 2)
					l.Debugln("Rule " + fmt.Sprint(i) + " is CRITICAL: " + rules.Rules[i].ToString())
					critical_fired += 1
				} else if actCount > rule.Warning {
					//r.StoreAlert(config.AlertInterval, rule.Pattern, "warning")
					icinga.SendResults(config, rule.Name, rule.Pattern, 1)
					l.Debugln("Rule " + fmt.Sprint(i) + " is WARNING: " + rules.Rules[i].ToString())
					warning_fired += 1
				} else {
					//r.StoreAlert(config.AlertInterval, rule.Pattern, "ok")
					icinga.SendResults(config, rule.Name, "Everythin is fine.", 0)
					l.Debugln("Rule " + fmt.Sprint(i) + " is OK: " + rules.Rules[i].ToString())
					ok_fired += 1
				}
			}
		}

		// check global counter 5 minutes
		c5 := r.GetGlobalCounter(5)
		// counter is above the defined limit => send icinga warning
		if c5 > rules.Global.FiveMinutes {
			icinga.SendResults(config, "Global 5m", "Global 5m", 1)
			l.Debugln("Global Rule 5m is WARNING")
		} else {
			icinga.SendResults(config, "Global 5m", "Global 5m", 0)
			l.Debugln("Global Rule 5m is OK")
		}
		l.Debugf("Gloabl counter %v minutes was at %v with limit at %v.", 5, c5, rules.Global.FiveMinutes)

		// check global counter 60 minutes
		c60 := r.GetGlobalCounter(60)
		// counter is above the defined limit => send icinga warning
		if c5 > rules.Global.FiveMinutes {
			icinga.SendResults(config, "Global 60m", "Global 60m", 1)
			l.Debugln("Global Rule 60m is WARNING")
		} else {
			icinga.SendResults(config, "Global 60m", "Global 60m", 0)
			l.Debugln("Global Rule 60m is OK")
		}
		l.Debugf("Gloabl counter %v minutes was at %v with limit at %v.", 60, c60, rules.Global.SixtyMinutes)

		l.Infof("Rule Status: OK:%v | WARNING:%v | CRITICAL:%v - Next run in %d seconds.", ok_fired, warning_fired, critical_fired, config.CheckInterval)
		time.Sleep(time.Duration(config.CheckInterval) * time.Second)
	}
}
