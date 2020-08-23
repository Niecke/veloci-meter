# Veloci-Meter

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