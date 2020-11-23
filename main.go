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
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/kardianos/service"
	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/background"
	"niecke-it.de/veloci-meter/config"
	"niecke-it.de/veloci-meter/mail"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
)

var logger service.Logger
var conf_path string
var log_path string

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
	config := config.LoadConfig(conf_path)

	f, err := os.OpenFile(log_path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	level, _ := l.ParseLevel(config.LogLevel)
	l.SetLevel(level)
	l.SetOutput(f)

	//##### RULES #####
	rules := rules.LoadRules()

	l.Infof(fmt.Sprint(len(rules.Rules)) + " rules loaded.")
	for i := 0; i < len(rules.Rules); i++ {
		l.Debugf("Rule " + fmt.Sprint(i) + ": " + rules.Rules[i].ToString())
	}

	//##### REDIS #####
	r := rdb.NewRDB(&config.Redis)

	// start the background process which checks key counts in redis
	//go background.CheckRedisLimits(config, rules)
	go background.CheckForAlerts(config, rules)

	//##### MAIL STUFF #####
	l.Infof("Check that mailboxes are setup...")
	imapClient := mail.NewIMAPClient(&config.Mail)

	// Login
	l.Infof("Loging into mail server...")
	if err := imapClient.Login(config.Mail.User, config.Mail.Password); err != nil {
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
		fetchMails(config, rules, r)
	}
}

func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.
	log.Println("I'm Stopping!")
	close(p.exit)
	return nil
}

func main() {
	if len(os.Args) == 3 {
		conf_path = os.Args[1]
		log_path = os.Args[1]
	} else {
		conf_path = "/opt/veloci-meter/config.json"
		log_path = "/var/log/veloci-meter.log"
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

func fetchMails(config *config.Config, rules *rules.Rules, r *rdb.RDBClient) {
	l.Debug("Running main process loop...")
	start_timestamp := int(time.Now().Unix())
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
			processed += 1
			for _, rule := range rules.Rules {
				if contains := strings.Contains(msg.Envelope.Subject, rule.Pattern); contains == true {
					r.StoreMail(msg, rule.Timeframe)
					known.AddNum(msg.SeqNum)
					found = true
					break
				}
			}
			if found == false {
				l.Debugln("Subject '" + msg.Envelope.Subject + "' does not match any pattern.")
				// increment the gloabl counters for unkown mails
				r.IncreaseGlobalCounter(5)
				l.Debugf("Increment gloabl counter %v minutes by 1.", 5)

				r.IncreaseGlobalCounter(60)
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

	end_timestamp := int(time.Now().Unix())
	duration := end_timestamp - start_timestamp
	l.Infof("%v of %v messages have been processed in %d seconds. Next run in %v seconds", processed, len(ids), duration, config.FetchInterval)
	time.Sleep(time.Duration(config.FetchInterval) * time.Second)
}
