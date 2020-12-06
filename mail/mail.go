package mail

import (
	"github.com/emersion/go-imap"
	move "github.com/emersion/go-imap-move"
	"github.com/emersion/go-imap/client"
	"niecke-it.de/veloci-meter/config"
	l "niecke-it.de/veloci-meter/logging"
)

// IMAPClient wraps a Client from github.com/emersion/go-imap/client and an extension from github.com/emersion/go-imap-move to directly move mails between two folders on the same server.
type IMAPClient struct {
	*client.Client
	*move.MoveClient
}

// NewIMAPClient connects to the mail server defined by the config and returns a pointer to the connected client.
func NewIMAPClient(conf *config.Mail) *IMAPClient {
	l.DebugLog("Connecting to mail server...", map[string]interface{}{
		"server_uri": conf.URI,
	})
	// Connect to server
	c, err := client.DialTLS(conf.URI, nil)
	if err != nil {
		l.ErrorLog(err, "There was an error while connecting to mail server.", map[string]interface{}{
			"server_uri": conf.URI,
		})
	}
	i := IMAPClient{
		c,
		move.NewClient(c),
	}
	l.DebugLog("Connecting to mail server successful.", map[string]interface{}{
		"server_uri": conf.URI,
	})

	return &i
}

// MarkAsSeen marks all mails in seSeq as seen. If there are no mails in setSeq the function returns immediately.
func (c *IMAPClient) MarkAsSeen(seqSet *imap.SeqSet) {
	if seqSet.Empty() != true {
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		if err := c.Store(seqSet, item, flags, nil); err != nil {
			l.ErrorLog(err, "IMAP Message Flag Update Failed", map[string]interface{}{
				"seq_set": seqSet,
			})
		}
		l.DebugLog("Mails flagged as seen.", map[string]interface{}{
			"mails": seqSet,
		})
	} else {
		l.DebugLog("No mails to flag as seen.", nil)
	}
}

// MoveToTODO moves all mails in seSeq to TODO. If there are no mails in setSeq the function returns immediately.
func (c *IMAPClient) MoveToTODO(seqSet *imap.SeqSet) {
	if seqSet.Empty() != true {
		if err := c.MoveWithFallback(seqSet, "ToDo"); err != nil {
			l.ErrorLog(err, "IMAP Message copy failed!", map[string]interface{}{
				"seq_set": seqSet,
			})
		}
		l.DebugLog("Mails moved to TODO.", map[string]interface{}{
			"mails": seqSet,
		})
	} else {
		l.DebugLog("No mails moved.", nil)
	}
}

// SearchUnseen returns a list of mails ids which are marked as unseen.
func (c *IMAPClient) SearchUnseen() []uint32 {
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := c.Search(criteria)
	if err != nil {
		l.ErrorLog(err, "There was an error while searching for unseen mails.", nil)
	}
	return ids
}
