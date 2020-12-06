package icinga

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"niecke-it.de/veloci-meter/config"
	l "niecke-it.de/veloci-meter/logging"
)

// SendResults send check data to the defined icinga server and logs a warning if no check definition was found on the server.
func SendResults(c *config.Config, name string, pattern string, exitCode int, count int64) {
	l.DebugLog("Sending results.", map[string]interface{}{
		"name":      name,
		"pattern":   pattern,
		"exit_code": exitCode,
	})
	// TODO move insecure ssl to config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: *c.InsecureSkipVerify},
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	var e = "OK"
	if exitCode == 1 {
		e = "WARNING"
	} else if exitCode == 2 {
		e = "CRITICAL"
	}
	var jsonStr = []byte(fmt.Sprintf(`{"type": "Service", "filter": "host.name==\"%v\" && service.name==\"%v\"", "exit_status": %d, "plugin_output": "[%v] Pattern: '%v'", "performance_data": [ "count=%d" ]}`, c.Icinga.Hostname, name, exitCode, e, pattern, count))
	resp, err := postForm(netClient, c.Icinga.Endpoint, c.Icinga.User, c.Icinga.Password, jsonStr)
	if err != nil {
		l.ErrorLog(err, "There was an error sending data to icinga.", map[string]interface{}{
			"payload": jsonStr,
		})
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.ErrorLog(err, "There was an error parsing answer from icinga.", map[string]interface{}{
			"body": resp.Body,
		})
	}
	l.DebugLog("Result from icinga.", map[string]interface{}{
		"body": body,
	})

	// TODO print Warning when result is empty
	var r map[string]interface{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		l.ErrorLog(err, "Error while decoding icinga result. The http response body was {{.body}}", map[string]interface{}{
			"body": body,
		})
	} else {
		results := r["results"].([]interface{})
		if len(results) == 0 {
			l.WarnLog("No Check definition found!", map[string]interface{}{
				"name":    name,
				"pattern": pattern,
			})
		} else {
			l.DebugLog("Send data for check {{.check}}", map[string]interface{}{
				"check": results[0],
			})
		}
	}

}

func postForm(c *http.Client, url string, user string, password string, data []byte) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(user, password)
	return c.Do(req)
}
