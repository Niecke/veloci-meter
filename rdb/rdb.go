package rdb

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/emersion/go-imap"
	"github.com/go-redis/redis"
	l "github.com/sirupsen/logrus"
	"niecke-it.de/veloci-meter/config"
)

// Client is the structure wrapping the redis client holding a connection to the server.
// The redis client can be directly accessed via Client.client
type Client struct {
	client *redis.Client
}

// NewClient uses the redis configuration provided to connect to a redis server and returns a pointer to the Client struct.
// TODO add reconnect with a wait of n seconds
func NewClient(c *config.Redis) *Client {
	l.WithFields(l.Fields{
		"Addr":       c.URI,
		"MaxRetries": 3,
		"Password":   "XXXX",
		"DB":         c.Database,
	}).Debug("Connect to redis...")
	r := Client{
		client: redis.NewClient(&redis.Options{
			Addr:       c.URI,
			MaxRetries: 3,
			Password:   c.Password, // no password set
			DB:         c.Database, // use default DB
		})}
	// Test the connection via ping
	if _, err := r.client.Ping().Result(); err != nil {
		l.Fatal(err)
	}
	l.Debug("Connection successful.")
	return &r
}

func buildHash(subject string) string {
	h := sha1.New()
	h.Write([]byte(subject))
	return hex.EncodeToString(h.Sum(nil))
}

// StoreMail takes the subject from the imap.Message and calculates the hash to store it in redis.
// To count multiple mails with the same subject an aditional random int32 is added to the redis key.
// TODO handel error while r.client.set()
func (r *Client) StoreMail(msg *imap.Message, duration int) {
	sha1Hash := buildHash(msg.Envelope.Subject)
	// using a random int32 as part of the redis key
	randomPart, _ := rand.Int(rand.Reader, big.NewInt(2147483647))
	r.client.Set(sha1Hash+":"+fmt.Sprint(randomPart), 1, time.Duration(duration)*time.Second)
	l.WithFields(l.Fields{
		"sha1Hash":   sha1Hash,
		"randomPart": randomPart.Text(10),
		"duration":   time.Duration(duration) * time.Second,
	}).Debugf("Stored %v:%v for %v", sha1Hash, randomPart.Text(10), time.Duration(duration)*time.Second)
}

// CountMail calls the redis eval function, to get all keys matching the provided pattern and then count the number of returned keys.
func (r *Client) CountMail(pattern string) int64 {
	sha1Hash := buildHash(pattern)
	v, err := r.client.Eval("return #redis.pcall('keys', '"+sha1Hash+":*')", nil).Result()
	if err != nil {
		l.Errorf("[%v] Error while counting mails in redis.", err)
		return int64(0)
	} else {
		l.WithFields(l.Fields{
			"mail_count": v.(int64),
			"pattern":    pattern,
		}).Debugf("There where %v mails for pattern '%v' in redis.", v.(int64), pattern)
		return v.(int64)
	}
}

func calculateGlobalKey(timestamp int, timeframe int) string {
	remainder := math.Mod(float64(timestamp), float64(timeframe*60))
	keyPart := timestamp - int(remainder)
	return "global:" + fmt.Sprint(timeframe) + ":" + fmt.Sprint(keyPart)
}

// timefram in minutes
func (r *Client) IncreaseGlobalCounter(timeframe int) {
	timestamp := int(time.Now().Unix())
	redisKey := calculateGlobalKey(timestamp, timeframe)
	err := r.client.Incr(redisKey)
	if err != nil {
		l.Debugf("[%v] Redis Command executed: [%v]", err, redisKey)
	}
}

func (r *Client) GetGlobalCounter(timeframe int) int {
	timestamp := int(time.Now().Unix())
	redisKey := calculateGlobalKey(timestamp, timeframe)
	val, err := r.client.Get(redisKey).Result()

	if err != nil {
		if err == redis.Nil {
			return 0
		}
		l.Errorf("[%v] There was an error while getting global counter from redis. Redis key was %v", err, redisKey)
	}

	c, err := strconv.Atoi(val)
	if err != nil {
		l.Errorf("[%v] There parsing global counter value from redis. value was %v", err, val)
	}

	return c
}

func (r *Client) GetKeys(pattern string) []string {
	val, err := r.client.Keys(pattern).Result()

	if err != nil {
		l.Errorf("[%v] There was an error while getting keys from redis. Key pattern was %v", err, pattern)
	}

	return val
}

func (r *Client) DeleteKey(key string) int64 {
	val, err := r.client.Del(key).Result()

	if err != nil {
		l.Errorf("[%v] There was an error while deleting %v from redis.", err, key)
	}

	return val
}
