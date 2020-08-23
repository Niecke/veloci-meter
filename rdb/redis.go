package rdb

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/go-redis/redis"
	"niecke-it.de/veloci-meter/config"
)

type rdb struct {
	client *redis.Client
}

func NewRDB(c *config.Redis) *rdb {
	log.Println("Connect to redis...")
	r := rdb{
		client: redis.NewClient(&redis.Options{
			Addr:       c.URI,
			MaxRetries: 3,
			Password:   c.Password, // no password set
			DB:         c.Database, // use default DB
		})}
	// Test redis
	if _, err := r.client.Ping().Result(); err != nil {
		log.Fatal(err)
	}
	log.Println("Connection successful.")
	return &r
}

func buildHash(subject string) string {
	h := sha1.New()
	h.Write([]byte(subject))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *rdb) StoreMail(msg *imap.Message, duration int) {
	sha1_hash := buildHash(msg.Envelope.Subject)
	r.client.Set(sha1_hash+":"+msg.Envelope.MessageId, msg.Envelope.MessageId, time.Duration(duration)*time.Second)
	log.Println("Stored " + sha1_hash + ":" + msg.Envelope.MessageId + " for " + fmt.Sprint(time.Duration(duration)*time.Second))
}

func (r *rdb) CountMail(pattern string) int64 {
	sha1_hash := buildHash(pattern)
	v, err := r.client.Eval("return #redis.pcall('keys', '"+sha1_hash+":*')", nil).Result()
	if err != nil {
		log.Fatal(err)
	}
	return v.(int64)
}
