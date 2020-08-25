package icinga

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/config"
)

func SendResults(c *config.Config, name string, pattern string, exitCode int) {
	l.Debugf("Sending results: name=%v | pattern=%v | exitCode=%v", name, pattern, exitCode)
	// TODO move insecure ssl to config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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
	var jsonStr = []byte(fmt.Sprintf(`{"type": "Service", "filter": "host.name==\"%v\" && service.name==\"%v\"", "exit_status": %d, "plugin_output": "[%v] %v"}`, c.Icinga.Hostname, name, exitCode, e, pattern))
	resp, err := PostForm(netClient, c.Icinga.Endpoint, c.Icinga.User, c.Icinga.Password, jsonStr)
	if err != nil {
		l.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Fatal(err)
	}
	l.Debugf("%s\n", string(body))
}

func PostForm(c *http.Client, url string, user string, password string, data []byte) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(user, password)
	return c.Do(req)
}
