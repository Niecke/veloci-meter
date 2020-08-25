package mail

import (
	"fmt"

	"github.com/emersion/go-imap"
	move "github.com/emersion/go-imap-move"
	"github.com/emersion/go-imap/client"
	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/config"
)

type IMAPClient struct {
	*client.Client
	*move.MoveClient
}

func NewIMAPClient(conf *config.Mail) *IMAPClient {
	l.Debugln("Connecting to mail server...")
	// Connect to server
	c, err := client.DialTLS(conf.URI, nil)
	if err != nil {
		l.Fatal(err)
	}
	i := IMAPClient{
		c,
		move.NewClient(c),
	}
	l.Debugln("Connected to mail server.")

	return &i
}

func (c *IMAPClient) MarkAsSeen(seqSet *imap.SeqSet) {
	if seqSet.Empty() != true {
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		if err := c.Store(seqSet, item, flags, nil); err != nil {
			l.Infoln("IMAP Message Flag Update Failed")
			l.Fatal(err)
		}
		l.Debugln("Mails flagged as seen: " + fmt.Sprint(seqSet))
	} else {
		l.Debugln("No mails to flag as seen.")
	}
}

func (c *IMAPClient) MoveToTODO(seqSet *imap.SeqSet) {
	if seqSet.Empty() != true {
		if err := c.MoveWithFallback(seqSet, "ToDo"); err != nil {
			l.Infoln("IMAP Message copy failed!")
			l.Fatal(err)
		}
		l.Debugln("Mails moved: " + fmt.Sprint(seqSet))
	} else {
		l.Debugln("No mails moved.")
	}
}

func (c *IMAPClient) SearchUnseen() []uint32 {
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := c.Search(criteria)
	if err != nil {
		l.Fatal(err)
	}
	return ids
}
