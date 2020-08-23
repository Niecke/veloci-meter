package main

/*
ToDo
* warn if no icinga check definition found
* add CleanUp run
* add reconnect for redis
* replace log.Fatal by Error Messages

*/

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"niecke-it.de/veloci-meter/background"
	"niecke-it.de/veloci-meter/config"
	"niecke-it.de/veloci-meter/mail"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
)

var ctx = context.Background()

func main() {
	//##### CONFIG #####
	config := config.LoadConfig()

	//##### RULES #####
	rules := rules.LoadRules()

	log.Println(fmt.Sprint(len(rules.Rules)) + " rules loaded.")
	for i := 0; i < len(rules.Rules); i++ {
		log.Println("Rule " + fmt.Sprint(i) + ": " + rules.Rules[i].ToString())
	}

	//##### REDIS #####
	r := rdb.NewRDB(&config.Redis)

	//##### SIGNALS #####
	// setup signal catching
	sigs := make(chan os.Signal, 1)

	// catch all signals since not explicitly listing
	signal.Notify(sigs)
	go func() {
		s := <-sigs
		log.Printf("RECEIVED SIGNAL: %s", s)
		if s == os.Interrupt || s == os.Kill {
			log.Println("Good Bye!")
		}
		os.Exit(1)
	}()

	// start the background process which checks key counts in redis
	//go background.CheckRedisLimits(config, rules)
	go background.CheckForAlerts(config, rules)

	//##### MAIL STUFF #####
	log.Printf("Check that mailboxes are setup...")
	imapClient := mail.NewIMAPClient(&config.Mail)

	// Login
	if err := imapClient.Login(config.Mail.User, config.Mail.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- imapClient.List("", "*", mailboxes)
	}()

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// check if todo mailbox exists
	if err := imapClient.Create("ToDo"); err == nil {
		log.Println("ToDo Mailbox was not present and was created.")
		imapClient.Subscribe("ToDo")
	}

	for {
		//##### MAIL STUFF #####
		imapClient := mail.NewIMAPClient(&config.Mail)

		// Login
		if err := imapClient.Login(config.Mail.User, config.Mail.Password); err != nil {
			log.Fatal(err)
		}
		log.Println("Logged in")

		// Select INBOX
		if _, err := imapClient.Select("INBOX", false); err != nil {
			log.Fatal(err)
		}

		//##### PROCESS MAILS #####
		// Set search criteria
		ids := imapClient.SearchUnseen()

		if len(ids) > 0 {
			log.Println(fmt.Sprint(len(ids)) + " new messages found. Processing max " + fmt.Sprint(config.BatchSize) + " of them.")
			unseenMails := new(imap.SeqSet)
			// only get the first n mails
			if len(ids) < config.BatchSize {
				unseenMails.AddNum(ids...)
			} else {
				unseenMails.AddNum(ids[:config.BatchSize]...)
			}

			messages := make(chan *imap.Message, 100)
			done := make(chan error, 1)
			go func() {
				done <- imapClient.Fetch(unseenMails, []imap.FetchItem{imap.FetchEnvelope}, messages)
			}()

			log.Println("Unseen messages:")
			unknown := new(imap.SeqSet)
			known := new(imap.SeqSet)
			for msg := range messages {
				found := false
				for _, rule := range rules.Rules {
					if contains := strings.Contains(msg.Envelope.Subject, rule.Pattern); contains == true {
						r.StoreMail(msg, rule.Timeframe)
						known.AddNum(msg.SeqNum)
						found = true
						break
					}
				}
				if found == false {
					log.Println("Subject '" + msg.Envelope.Subject + "' does not match any pattern.")
					unknown.AddNum(msg.SeqNum)
				}
			}
			imapClient.MarkAsSeen(known)
			imapClient.MoveToTODO(unknown)

			if err := <-done; err != nil {
				log.Fatal(err)
			}
		} else {
			log.Println("No new messages found.")
		}

		imapClient.Logout()
		imapClient.Terminate()

		log.Printf("Sleep for %ds before fetching mails again.", config.FetchInterval)
		time.Sleep(time.Duration(config.FetchInterval) * time.Second)
	}
}
