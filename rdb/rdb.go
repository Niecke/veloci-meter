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
		l.WithFields(l.Fields{
			"error": err,
		}).Errorf("[%v] Error while counting mails in redis.", err)
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

// IncreaseGlobalCounter increments the global counter for the provided timeframe in minutes.
// If the timeframe for example is 5 (minutes) then the function calculates the redis key for this timeframe based on the actual time and increments this counter.
// If there is no redis key it will be 1 after this operation.
func (r *Client) IncreaseGlobalCounter(timeframe int) {
	timestamp := int(time.Now().Unix())
	redisKey := calculateGlobalKey(timestamp, timeframe)
	val, err := r.client.Incr(redisKey).Result()
	if err != nil {
		l.WithFields(l.Fields{
			"error":     err,
			"redis_key": redisKey,
		}).Errorf("[%v] Redis Command executed: [%v]", err, redisKey)
	} else {
		l.WithFields(l.Fields{
			"timeframe":    timeframe,
			"redis_result": val,
		}).Debugf("The global counter for timeframe %v minutes was increased.", timeframe)
	}
}

// GetGlobalCounter returns the number of mails for the actual timeframe of n minutes.
// If there has been no data in redis this function returns 0 since there have been no mails processed for this timestamp.
func (r *Client) GetGlobalCounter(timeframe int) int {
	timestamp := int(time.Now().Unix())
	redisKey := calculateGlobalKey(timestamp, timeframe)
	val, err := r.client.Get(redisKey).Result()

	if err != nil {
		if err == redis.Nil {
			// If err == redis.Nil there has been no redis key for the actual timeframe thus no mails have been processed
			l.WithFields(l.Fields{
				"timeframe":    timeframe,
				"redis_result": 0,
			}).Debugf("There has been no data for timefram '%v' minutes thus 0 was returned.", timeframe)
			return 0
		}
		l.WithFields(l.Fields{
			"error":     err,
			"redis_key": redisKey,
		}).Errorf("[%v] There was an error while getting global counter from redis. Redis key was %v", err, redisKey)
	}

	// Parse the result from redis into int
	c, err := strconv.Atoi(val)
	if err != nil {
		l.WithFields(l.Fields{
			"error":        err,
			"redis_result": val,
		}).Errorf("[%v] There parsing global counter value from redis. value was %v", err, val)
	}

	l.WithFields(l.Fields{
		"timeframe":    timeframe,
		"redis_result": c,
	}).Debugf("There have been %d mails for timefram '%v' minutes.", c, timeframe)
	return c
}

func (r *Client) GetKeys(pattern string) []string {
	val, err := r.client.Keys(pattern).Result()

	if err != nil {
		l.WithFields(l.Fields{
			"error":   err,
			"pattern": pattern,
		}).Errorf("[%v] There was an error while getting keys from redis. Key pattern was %v", err, pattern)
		return []string{}
	} else {
		l.WithFields(l.Fields{
			"pattern":      pattern,
			"redis_result": val,
		}).Debugf("There has been %d keys for pattern '%v'", len(val), pattern)
		return val
	}
}

func (r *Client) DeleteKey(key string) int64 {
	val, err := r.client.Del(key).Result()

	if err != nil {
		l.WithFields(l.Fields{
			"error":     err,
			"redis_key": key,
		}).Errorf("[%v] There was an error while deleting %v from redis.", err, key)
	}

	return val
}
