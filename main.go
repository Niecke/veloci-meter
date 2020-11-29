package main

/*
ToDo
* warn if no icinga check definition found
* add CleanUp run
* add reconnect for redis
* replace log.Fatal by Error Messages

*/

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/kardianos/service"
	"github.com/robfig/cron/v3"
	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/background"
	"niecke-it.de/veloci-meter/config"
	"niecke-it.de/veloci-meter/mail"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
)

var logger service.Logger
var confPath string
var logPath string
var conf config.Config
var cronJob cron.Cron

// Program structures.
//  Define Start and Stop methods.
type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	if service.Interactive() {
		log.Println("Running in terminal.")
	} else {
		log.Println("Running under service manager.")
	}
	p.exit = make(chan struct{})

	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() error {
	//##### CONFIG #####
	conf = *config.LoadConfig(confPath)

	f, err := os.OpenFile(logPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	level, _ := l.ParseLevel(conf.LogLevel)
	l.SetLevel(level)
	l.SetOutput(f)
	if conf.LogFormat == "JSON" {
		l.SetFormatter(&l.JSONFormatter{})
	}

	//##### RULES #####
	rules := rules.LoadRules()

	l.Infof(fmt.Sprint(len(rules.Rules)) + " rules loaded.")
	for i := 0; i < len(rules.Rules); i++ {
		l.Debugf("Rule " + fmt.Sprint(i) + ": " + rules.Rules[i].ToString())
	}

	//##### CRON #####
	cronJob = *cron.New()
	cronJob.AddFunc(conf.CleanUpSchedule, cleanUp)
	l.Infof("Cron cleanup job started with '%v' schedule.", conf.CleanUpSchedule)
	cronJob.Start()

	//##### REDIS #####
	r := rdb.NewClient(&conf.Redis)

	// start the background process which checks key counts in redis
	//go background.CheckRedisLimits(config, rules)
	go background.CheckForAlerts(&conf, rules)

	//##### MAIL STUFF #####
	l.Infof("Check that mailboxes are setup...")
	imapClient := mail.NewIMAPClient(&conf.Mail)

	// Login
	l.Infof("Loging into mail server...")
	if err := imapClient.Login(conf.Mail.User, conf.Mail.Password); err != nil {
		l.Fatal(err)
	}

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- imapClient.List("", "*", mailboxes)
	}()

	if err := <-done; err != nil {
		l.Fatal(err)
	}

	// check if todo mailbox exists
	if err := imapClient.Create("ToDo"); err == nil {
		l.Infoln("'ToDo' Mailbox was not present and was created.")
		imapClient.Subscribe("ToDo")
	}

	for {
		fetchMails(&conf, rules, r)
	}
}

func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.

	l.Infof("Cron cleanup job stopping...")
	channel := cronJob.Stop().Done()
	waitForChannelsToClose(channel)
	l.Infof("Cron cleanup job stoped!")

	l.Infof("Service stopping!")
	close(p.exit)
	return nil
}

func waitForChannelsToClose(ch <-chan struct{}) {
	t := time.Now()
	l.Debugf("%v for Cron job to stop", time.Since(t))
}

var GlobalPatterns = map[string]string{
	"5m":  "global:5:",
	"60m": "global:60:",
}

func cleanUp() {
	l.Debug("Running clean up job.")
	timestamp := int(time.Now().Unix())
	deletedKey := 0
	l.Debug("Connecting to redis...")
	r := rdb.NewClient(&conf.Redis)

	for index, val := range GlobalPatterns {
		l.Debugf("Checking %v keys...", index)
		keys := r.GetKeys(val + "*")
		for _, key := range keys {
			ts, err := strconv.Atoi(strings.Replace(key, val, "", -1))
			if err != nil {
				l.Errorf("[%v] There was an error converting %v to int.", err, key)
			} else {
				// if the key is older than 24 hours -> delete it
				if timestamp-ts > 86400 {
					redisReturn := r.DeleteKey(key)
					l.Infof("Redis return for deleting %v was %v", key, redisReturn)
					deletedKey++
				}
			}
		}
	}
	end := int(time.Now().Unix())
	duration := end - timestamp
	l.Infof("Cleanup job is done. Deleted %v keys from redis in %v seconds.", deletedKey, duration)
}

func main() {
	if len(os.Args) == 3 {
		confPath = os.Args[1]
		logPath = os.Args[1]
	} else {
		confPath = "/opt/veloci-meter/config.json"
		logPath = "/var/log/veloci-meter.log"
	}

	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	svcConfig := &service.Config{
		Name:        "GoServiceExampleLogging",
		DisplayName: "Go Service Example for Logging",
		Description: "This is an example Go service that outputs log messages.",
		Dependencies: []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"},
		Option: options,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

func fetchMails(config *config.Config, rules *rules.Rules, r *rdb.Client) {
	l.Debug("Running main process loop...")
	startTimestamp := int(time.Now().Unix())
	//##### MAIL STUFF #####
	imapClient := mail.NewIMAPClient(&config.Mail)

	// Login
	if err := imapClient.Login(config.Mail.User, config.Mail.Password); err != nil {
		l.Fatal(err)
	}
	l.Debugln("Logged in")

	// Select INBOX
	if _, err := imapClient.Select("INBOX", false); err != nil {
		l.Fatal(err)
	}

	//##### PROCESS MAILS #####
	// Set search criteria
	ids := imapClient.SearchUnseen()
	processed := 0
	if len(ids) > 0 {
		l.Debugln(fmt.Sprint(len(ids)) + " new messages found. Processing max " + fmt.Sprint(config.Mail.BatchSize) + " of them.")
		unseenMails := new(imap.SeqSet)
		// only get the first n mails
		if len(ids) < config.Mail.BatchSize {
			unseenMails.AddNum(ids...)
		} else {
			unseenMails.AddNum(ids[:config.Mail.BatchSize]...)
		}

		messages := make(chan *imap.Message, 100)
		done := make(chan error, 1)
		l.Debugf("%v", unseenMails)
		go func() {
			done <- imapClient.Fetch(unseenMails, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		l.Debugln("There are unseen messages.")
		unknown := new(imap.SeqSet)
		known := new(imap.SeqSet)

		for msg := range messages {
			found := false
			processed++
			for _, rule := range rules.Rules {
				if contains := strings.Contains(msg.Envelope.Subject, rule.Pattern); contains == true {
					r.StoreMail(msg, rule.Timeframe)
					r.IncreaseStatisticCountMail(rule.Name)
					known.AddNum(msg.SeqNum)
					found = true
					break
				}
			}
			if found == false {
				l.Debugln("Subject '" + msg.Envelope.Subject + "' does not match any pattern.")
				// increment the gloabl counters for unkown mails
				r.IncreaseGlobalCounter(5)
				r.IncreaseStatisticCountMail("Global 5m")
				l.Debugf("Increment gloabl counter %v minutes by 1.", 5)

				r.IncreaseGlobalCounter(60)
				r.IncreaseStatisticCountMail("Global 60m")
				l.Debugf("Increment gloabl counter %v minutes by 1.", 60)
				unknown.AddNum(msg.SeqNum)
			}
		}
		imapClient.MarkAsSeen(known)
		imapClient.MoveToTODO(unknown)

		if err := <-done; err != nil {
			l.Fatal(err)
		}
	} else {
		l.Debugln("No new messages found.")
	}

	imapClient.Logout()
	imapClient.Terminate()

	endTimestamp := int(time.Now().Unix())
	duration := endTimestamp - startTimestamp
	l.Infof("%v of %v messages have been processed in %d seconds. Next run in %v seconds", processed, len(ids), duration, config.FetchInterval)
	time.Sleep(time.Duration(config.FetchInterval) * time.Second)
}
