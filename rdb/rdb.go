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

type Stats struct {
	Name     string
	Mail     int64
	Warning  int64
	Critical int64
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

func (r *Client) Client() *redis.Client {
	return r.client
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
	}

	l.WithFields(l.Fields{
		"mail_count": v.(int64),
		"pattern":    pattern,
	}).Debugf("There where %v mails for pattern '%v' in redis.", v.(int64), pattern)
	return v.(int64)
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
		return 0
	}

	// Parse the result from redis into int
	c, err := strconv.Atoi(val)
	if err != nil {
		l.WithFields(l.Fields{
			"error":        err,
			"redis_result": val,
		}).Errorf("[%v] There was an error while parsing global counter value from redis. value was %v", err, val)
	}

	l.WithFields(l.Fields{
		"timeframe":    timeframe,
		"redis_result": c,
	}).Debugf("There have been %d mails for timefram '%v' minutes.", c, timeframe)
	return c
}

// GetKeys calls the redis keys command with the specified pattern and returns list of matching keys.
func (r *Client) GetKeys(pattern string) []string {
	val, err := r.client.Keys(pattern).Result()

	if err != nil {
		l.WithFields(l.Fields{
			"error":   err,
			"pattern": pattern,
		}).Errorf("[%v] There was an error while getting keys from redis. Key pattern was %v", err, pattern)
		return []string{}
	}

	l.WithFields(l.Fields{
		"pattern":      pattern,
		"redis_result": val,
	}).Debugf("There has been %d keys for pattern '%v'", len(val), pattern)
	return val

}

// DeleteKey calls the redis del function and returns the result.
// If there was an error zero is returned.
func (r *Client) DeleteKey(key string) int64 {
	val, err := r.client.Del(key).Result()

	if err != nil {
		l.WithFields(l.Fields{
			"error":     err,
			"redis_key": key,
		}).Errorf("[%v] There was an error while deleting %v from redis.", err, key)
		return 0
	}

	l.WithFields(l.Fields{
		"redis_key":    key,
		"redis_result": val,
	}).Debugf("The key %v has been deleted", key)
	return val
}

func (r *Client) increaseStatisticCount(name string, t string) int64 {
	ts := int(time.Now().Unix())
	timestampDay := ts - int(math.Mod(float64(ts), float64(24*60*60)))
	val, err := r.client.HIncrBy("stats:"+name+":"+fmt.Sprint(timestampDay), t, int64(1)).Result()

	if err != nil {
		l.WithFields(l.Fields{
			"error":     err,
			"redis_key": name,
		}).Errorf("[%v] There was an error while increasing stats [%v] for name '%v'.", err, t, name)
		return int64(0)
	}

	l.WithFields(l.Fields{
		"redis_key":    name,
		"redis_result": val,
	}).Debugf("Stats counter [%v] for name '%v' is now at %v", t, name, val)
	return val
}

// IncreaseStatisticCountMail is used to count mails per name without epiring the count.
// This is used to check if a rules has any hits.
func (r *Client) IncreaseStatisticCountMail(name string) int64 {
	return r.increaseStatisticCount(name, "mail")
}

func (r *Client) IncreaseStatisticCountWarning(name string) int64 {
	return r.increaseStatisticCount(name, "warning")
}

func (r *Client) IncreaseStatisticCountCritical(name string) int64 {
	return r.increaseStatisticCount(name, "critical")
}

// GetStatisticCount returns the actual number of hits for one name.
func (r *Client) GetStatisticCount(name string, timestamp int) Stats {
	stats := Stats{Name: name, Mail: 0, Warning: 0, Critical: 0}
	timestampDay := timestamp - int(math.Mod(float64(timestamp), float64(24*60*60)))
	val, err := r.client.HMGet("stats:"+name+":"+fmt.Sprint(timestampDay), "mail", "warning", "critical").Result()

	if err != nil {
		l.WithFields(l.Fields{
			"error":     err,
			"redis_key": name,
			"stats":     stats,
		}).Errorf("[%v] There was an error while getting stats for name '%v'.", err, name)
		return stats
	}

	if val[0] != nil {
		result, err := strconv.ParseInt(fmt.Sprint(val[0]), 10, 64)
		if err != nil {
			l.WithFields(l.Fields{
				"error":        err,
				"stats_type":   "mail",
				"redis_result": val,
				"stats":        stats,
			}).Errorf("[%v] There was an error while parsing stats counter value [%v] for name '%v' from redis. value was %v", err, "mail", name, val[0])
		} else {
			stats.Mail = result
		}
	} else {
		stats.Mail = 0
	}

	if val[1] != nil {
		result, err := strconv.ParseInt(fmt.Sprint(val[1]), 10, 64)
		if err != nil {
			l.WithFields(l.Fields{
				"error":        err,
				"stats_type":   "warning",
				"redis_result": val,
				"stats":        stats,
			}).Errorf("[%v] There was an error while parsing stats counter value [%v] for name '%v' from redis. value was %v", err, "warning", name, val[1])
		} else {
			stats.Warning = result
		}
	} else {
		stats.Warning = 0
	}

	if val[2] != nil {
		result, err := strconv.ParseInt(fmt.Sprint(val[2]), 10, 64)
		if err != nil {
			l.WithFields(l.Fields{
				"error":        err,
				"stats_type":   "critical",
				"redis_result": val,
				"stats":        stats,
			}).Errorf("[%v] There was an error while parsing stats counter value [%v] for name '%v' from redis. value was %v", err, "critical", name, val[2])
		} else {
			stats.Critical = result
		}
	} else {
		stats.Critical = 0
	}

	l.WithFields(l.Fields{
		"redis_key":    name,
		"redis_result": val,
		"stats":        stats,
	}).Debugf("The stats counter for name '%v' is at %v", name, stats)

	return stats
}
