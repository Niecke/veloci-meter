package mail

import (
	"fmt"
	"log"

	"github.com/emersion/go-imap"
	move "github.com/emersion/go-imap-move"
	"github.com/emersion/go-imap/client"
	"niecke-it.de/veloci-meter/config"
)

type IMAPClient struct {
	*client.Client
	*move.MoveClient
}

func NewIMAPClient(conf *config.Mail) *IMAPClient {
	log.Println("Connecting to server...")
	// Connect to server
	c, err := client.DialTLS(conf.URI, nil)
	if err != nil {
		log.Fatal(err)
	}
	i := IMAPClient{
		c,
		move.NewClient(c),
	}
	log.Println("Connected")

	return &i
}

func (c *IMAPClient) MarkAsSeen(seqSet *imap.SeqSet) {
	if seqSet.Empty() != true {
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		if err := c.Store(seqSet, item, flags, nil); err != nil {
			log.Println("IMAP Message Flag Update Failed")
			log.Fatal(err)
		}
		log.Println("Mails flagged as seen: " + fmt.Sprint(seqSet))
	} else {
		log.Println("No mails to flagg as seen.")
	}
}

func (c *IMAPClient) MoveToTODO(seqSet *imap.SeqSet) {
	if seqSet.Empty() != true {
		if err := c.MoveWithFallback(seqSet, "ToDo"); err != nil {
			log.Println("IMAP Message copy failed!")
			log.Fatal(err)
		}
		log.Println("Mails moved: " + fmt.Sprint(seqSet))
	} else {
		log.Println("No mails moved.")
	}
}

func (c *IMAPClient) SearchUnseen() []uint32 {
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}
	return ids
}
