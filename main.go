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
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/kardianos/service"
	"github.com/robfig/cron/v3"
	"niecke-it.de/veloci-meter/background"
	"niecke-it.de/veloci-meter/config"
	l "niecke-it.de/veloci-meter/logging"
	"niecke-it.de/veloci-meter/mail"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
	"niecke-it.de/veloci-meter/stats"
)

var logger service.Logger
var confPath string
var logPath string
var conf config.Config
var rulesList rules.Rules
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

	//##### LOGGER #####
	l.SetUpLogger(logPath, conf.LogLevel, conf.LogFormat)

	//##### RULES #####
	rulesList := rules.LoadRules("/opt/veloci-meter/rules.json")

	l.InfoLog("{{.rule_count}} rules loaded.", map[string]interface{}{"rule_count": len(rulesList.Rules)})
	for i := 0; i < len(rulesList.Rules); i++ {
		l.DebugLog("Rule {{.rule_position}}: {{.rule_content}}", map[string]interface{}{"rule_position": i, "rule_content": rulesList.Rules[i].ToString()})
	}

	//##### CRON #####
	cronJob = *cron.New()
	cronJob.AddFunc(conf.CleanUpSchedule, cleanUp)
	l.InfoLog("Cron job [{{.job_name}}] started with '{{.job_schedule}}' schedule.", map[string]interface{}{"job_name": "cleanUp", "job_schedule": conf.CleanUpSchedule})
	exportJobSchedule := "0 3 * * *"
	cronJob.AddFunc(exportJobSchedule, wrapExportJob)
	l.InfoLog("Cron job [{{job_name}}] started with '{{job_schedule}}' schedule.", map[string]interface{}{"job_name": "exportJob", "job_schedule": exportJobSchedule})
	cronJob.Start()

	//##### REDIS #####
	r := rdb.NewClient(&conf.Redis)

	// start the background process which checks key counts in redis
	//go background.CheckRedisLimits(config, rules)
	go background.CheckForAlerts(&conf, rulesList)

	//##### MAIL STUFF #####
	l.InfoLog("Check that mailboxes are setup...", nil)
	imapClient := mail.NewIMAPClient(&conf.Mail)

	// Login
	l.InfoLog("Loging into mail server...", nil)
	if err := imapClient.Login(conf.Mail.User, conf.Mail.Password); err != nil {
		l.FatalLog(err, "Error while logging into mailbox.", map[string]interface{}{"user": conf.Mail.User})
	}

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- imapClient.List("", "*", mailboxes)
	}()

	if err := <-done; err != nil {
		l.FatalLog(err, "Error while reading mails from the server.", map[string]interface{}{"user": conf.Mail.User})
	}

	// check if todo mailbox exists
	if err := imapClient.Create("ToDo"); err == nil {
		l.InfoLog("'ToDo' Mailbox was not present and was created.", nil)
		imapClient.Subscribe("ToDo")
	}

	for {
		fetchMails(&conf, rulesList, r)
	}
}

func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.
	l.InfoLog("Cron jobs stopping...", nil)
	channel := cronJob.Stop().Done()
	waitForChannelsToClose(channel)
	l.InfoLog("Cron jobs stopped!", nil)

	l.InfoLog("Service stopping!", nil)
	close(p.exit)
	return nil
}

func waitForChannelsToClose(ch <-chan struct{}) {
	t := time.Now()
	l.DebugLog("{{.duration}} for Cron job to stop", map[string]interface{}{"duration": time.Since(t)})
}

var GlobalPatterns = map[string]string{
	"5m":  "global:5:",
	"60m": "global:60:",
}

func wrapExportJob() {
	stats.ExportJob(&conf, rulesList)
}

func cleanUp() {
	l.DebugLog("Running clean up job.", nil)
	timestamp := int(time.Now().Unix())
	deletedKey := 0
	l.DebugLog("Connecting to redis...", map[string]interface{}{"redis_uri": conf.Redis.URI})
	r := rdb.NewClient(&conf.Redis)

	for index, val := range GlobalPatterns {
		l.DebugLog("Checking {{.index}} keys...", map[string]interface{}{"index": index})
		keys := r.GetKeys(val + "*")
		for _, key := range keys {
			ts, err := strconv.Atoi(strings.Replace(key, val, "", -1))
			if err != nil {
				l.ErrorLog(err, "There was an error converting {{.data}} to int.", map[string]interface{}{"data": key})
			} else {
				// if the key is older than 24 hours -> delete it
				if timestamp-ts > 86400 {
					redisReturn := r.DeleteKey(key)
					l.InfoLog("Redis return for deleting {{.redis_key}} was {{.redis_result}}", map[string]interface{}{"redis_key": key, "redis_result": redisReturn})
					deletedKey++
				}
			}
		}
	}
	end := int(time.Now().Unix())
	duration := end - timestamp
	l.InfoLog("Cleanup job is done. Deleted {{.redis_key}} keys from redis in {{.duration}} seconds.", map[string]interface{}{"redis_key": deletedKey, "duration": duration})
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
	l.DebugLog("Running main process loop...", nil)
	startTimestamp := int(time.Now().Unix())
	//##### MAIL STUFF #####
	imapClient := mail.NewIMAPClient(&config.Mail)

	// Login
	if err := imapClient.Login(config.Mail.User, config.Mail.Password); err != nil {
		l.FatalLog(err, "There was an error while logging into mailbox.", map[string]interface{}{"mail_user": config.Mail.User})
	}
	l.DebugLog("Logged in", nil)

	// Select INBOX
	if _, err := imapClient.Select("INBOX", false); err != nil {
		l.FatalLog(err, "There was an error selecting INBOX.", nil)
	}

	//##### PROCESS MAILS #####
	// Set search criteria
	ids := imapClient.SearchUnseen()
	processed := 0
	if len(ids) > 0 {
		l.DebugLog("{{.count}} new messages found. Processing max {{.batch_size}} of them.", map[string]interface{}{"count": len(ids), "batch_size": config.Mail.BatchSize})
		unseenMails := new(imap.SeqSet)
		// only get the first n mails
		if len(ids) < config.Mail.BatchSize {
			unseenMails.AddNum(ids...)
		} else {
			unseenMails.AddNum(ids[:config.Mail.BatchSize]...)
		}

		messages := make(chan *imap.Message, 100)
		done := make(chan error, 1)
		l.DebugLog("Messages will be processed.", map[string]interface{}{"unseen_mails": unseenMails})
		go func() {
			done <- imapClient.Fetch(unseenMails, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

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
				l.DebugLog("Subject '{{.message_subject}}' does not match any pattern.", map[string]interface{}{"message_subject": msg.Envelope.Subject})
				// increment the gloabl counters for unkown mails
				r.IncreaseGlobalCounter(5)
				r.IncreaseStatisticCountMail("Global 5m")
				l.DebugLog("Increment gloabl counter 5 minutes by 1.", nil)

				r.IncreaseGlobalCounter(60)
				r.IncreaseStatisticCountMail("Global 60m")
				l.DebugLog("Increment gloabl counter 60 minutes by 1.", nil)
				unknown.AddNum(msg.SeqNum)
			}
		}
		imapClient.MarkAsSeen(known)
		imapClient.MoveToTODO(unknown)

		if err := <-done; err != nil {
			l.FatalLog(err, "Unkown error!", nil)
		}
	} else {
		l.DebugLog("No new messages found.", nil)
	}

	imapClient.Logout()
	imapClient.Terminate()

	endTimestamp := int(time.Now().Unix())
	duration := endTimestamp - startTimestamp
	l.InfoLog("{{.processed}} of {{.count}} messages have been processed in {{.duration}} seconds. Next run in {{.fetch_interval}} seconds", map[string]interface{}{"processed": processed, "count": len(ids), "duration": duration, "fetch_interval": conf.FetchInterval})
	time.Sleep(time.Duration(config.FetchInterval) * time.Second)
}
