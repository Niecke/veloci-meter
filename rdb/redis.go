package rdb

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/go-redis/redis"
	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/config"
)

type RDBClient struct {
	client *redis.Client
}

func NewRDB(c *config.Redis) *RDBClient {
	l.Debugln("Connect to redis...")
	r := RDBClient{
		client: redis.NewClient(&redis.Options{
			Addr:       c.URI,
			MaxRetries: 3,
			Password:   c.Password, // no password set
			DB:         c.Database, // use default DB
		})}
	// Test redis
	if _, err := r.client.Ping().Result(); err != nil {
		l.Fatal(err)
	}
	l.Debugln("Connection successful.")
	return &r
}

func buildHash(subject string) string {
	h := sha1.New()
	h.Write([]byte(subject))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *RDBClient) StoreMail(msg *imap.Message, duration int) {
	sha1_hash := buildHash(msg.Envelope.Subject)
	r.client.Set(sha1_hash+":"+msg.Envelope.MessageId, msg.Envelope.MessageId, time.Duration(duration)*time.Second)
	l.Debugln("Stored " + sha1_hash + ":" + msg.Envelope.MessageId + " for " + fmt.Sprint(time.Duration(duration)*time.Second))
}

func (r *RDBClient) CountMail(pattern string) int64 {
	sha1_hash := buildHash(pattern)
	v, err := r.client.Eval("return #redis.pcall('keys', '"+sha1_hash+":*')", nil).Result()
	if err != nil {
		l.Fatal(err)
	}
	return v.(int64)
}
