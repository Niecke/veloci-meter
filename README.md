# Veloci-Meter

## Missing Features

* [ ] refactor the main.go
* [ ] single check mode
* [x] enhance logging (add debug log)
* [x] run programm as service
* [ ] warn if icinga checks are not defined
* [ ] export icinga check configuration when starting

## Installation

The binary `veloci-meter` and both json files `rules.json` and `config.json` need to be placed in `/opt/veloci-meter`.
For running the veloci-meter as a systemd service the `veloci-meter.service` needs to be place in `/etc/systemd/system/`

## Rules

The file `rules.example.json` should be renamed to `rules.json`. 
This file containts three example rules. 
The first checks for mails containing the string `Service-Mail` in the mail subject. 
If there is one or more mails of this type in the last 10 seconds, a warning will be send to icinga.
If there are 5 or more mails in the last 10 seconds a critical alert will be send to icinga. 
The second rule will always be critical an never warning since the value for warning is greater than the value for critical.
The last rule is one that will be warning, if there is no mail with a subject containing `Hello` in the last hour.
For this rule you can also define `"alert":"critical"` so the rule won't be warning but critical if is less than one mail each hour.
```json
{
    "rules": [
        {
            "name": "first rule",
            "pattern": "Service-Mail",
            "timeframe": 10,
            "warning": 1,
            "critical": 5
        },
        {
            "name": "another rule",
            "pattern": "Test",
            "timeframe": 10,
            "warning": 3,
            "critical": 1
        },
        {
            "name": "positive rule",
            "pattern": "Hello",
            "timeframe": 3600,
            "ok": 1
        }
    ]
}
```

### Global Rules

There are also rules which apply to all mails which do not map any pattern.
The limits for these mails must be defined within the `rules.json` file as follow:
```
{
    "global": {
        "5": 10,
        "60": 50
    },
    "rules": [...]
}
```
There is only one rule for the timeframe of 5 minutes and one for 60 minutes. If the defined limits are reached a warning will be send to icinga.

## Config

- `Mail.BatchSize` The number of mails processed within one iteration.
- `FetchIntervanl` The number of seonds waited before fetching mails again.
- `CheckIntervanl` The number of seonds waited data in redis is check again and notifications are send to icinga.

## Icinga2 Config

Add something like the following to the icinga2 config directory place in `/etc/icinga2/conf.d`

```txt
object Host "MAIL" {
  address = "10.0.0.1"
  check_command = "hostalive"
}

object Service "first rule" {
  host_name = "MAIL"
  check_command = "dummy"
  check_command = "passive"
  max_check_attempts = "1"
  enable_active_checks = false
  enable_passive_checks = true
}

object Service "another rule" {
  host_name = "MAIL"
  check_command = "dummy"
  check_command = "passive"
  max_check_attempts = "1"
  enable_active_checks = false
  enable_passive_checks = true
}

object Service "positive rule" {
  host_name = "MAIL"
  check_command = "dummy"
  check_command = "passive"
  max_check_attempts = "1"
  enable_active_checks = false
  enable_passive_checks = true
}
```
