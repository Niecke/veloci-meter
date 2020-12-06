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
	"niecke-it.de/veloci-meter/config"
	l "niecke-it.de/veloci-meter/logging"
)

// Client is the structure wrapping the redis client holding a connection to the server.
// The redis client can be directly accessed via Client.client
type Client struct {
	client *redis.Client
}

// Stats is the internal structure for storing counts per rule. It contains the name of a rule and counter for for mails matching this rule.
// In addition the number of warning and critical alerts is stored in this struct.
type Stats struct {
	Name     string
	Mail     int64
	Warning  int64
	Critical int64
}

// NewClient uses the redis configuration provided to connect to a redis server and returns a pointer to the Client struct.
// TODO add reconnect with a wait of n seconds
func NewClient(c *config.Redis) *Client {
	l.DebugLog("Connect to redis...", map[string]interface{}{
		"Addr":       c.URI,
		"MaxRetries": 3,
		"Password":   "XXXX",
		"DB":         c.Database})
	r := Client{
		client: redis.NewClient(&redis.Options{
			Addr:       c.URI,
			MaxRetries: 3,
			Password:   c.Password, // no password set
			DB:         c.Database, // use default DB
		})}
	// Test the connection via ping
	if _, err := r.client.Ping().Result(); err != nil {
		l.FatalLog(err, "Unknown error", nil)
	}
	l.DebugLog("Connection successful.", nil)
	return &r
}

// Client is only used to acess the internal Redis client from out of the rdb package.
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
	l.DebugLog("Stored {{.sha1_hash}}:{{.random_part}} for {{.duration}}", map[string]interface{}{
		"sha1_hash":   sha1Hash,
		"random_part": randomPart.Text(10),
		"duration":    time.Duration(duration) * time.Second})
}

// CountMail calls the redis eval function, to get all keys matching the provided pattern and then count the number of returned keys.
func (r *Client) CountMail(pattern string) int64 {
	sha1Hash := buildHash(pattern)
	v, err := r.client.Eval("return #redis.pcall('keys', '"+sha1Hash+":*')", nil).Result()
	if err != nil {
		l.ErrorLog(err, "Error while counting mails in redis.", nil)
		return int64(0)
	}

	l.DebugLog("There where {{.mail_count}} mails for pattern '{{.pattern}}' in redis.", map[string]interface{}{
		"mail_count": v.(int64),
		"pattern":    pattern})
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
		l.ErrorLog(err, "Redis Command executed: [INCR {{.redis_key}}]", map[string]interface{}{
			"redis_key": redisKey,
		})
	} else {
		l.DebugLog("The global counter for timeframe {{.timeframe}} minutes was increased.", map[string]interface{}{
			"timeframe":    timeframe,
			"redis_result": val,
		})
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
			l.DebugLog("There has been no data for timefram {{.timeframe}} minutes thus 0 was returned.", map[string]interface{}{
				"timeframe":    timeframe,
				"redis_result": 0})
			return 0
		}
		l.ErrorLog(err, "There was an error while getting global counter from redis.", map[string]interface{}{
			"redis_key": redisKey,
		})
		return 0
	}

	// Parse the result from redis into int
	c, err := strconv.Atoi(val)
	if err != nil {
		l.ErrorLog(err, "There was an error while parsing global counter value from redis. value was {{.redis_result}}", map[string]interface{}{
			"redis_result": val})
	}

	l.DebugLog("There have been {{.redis_result}} mails for timeframe '{{.timeframe}}' minutes.", map[string]interface{}{
		"timeframe":    timeframe,
		"redis_result": c,
	})
	return c
}

// GetKeys calls the redis keys command with the specified pattern and returns list of matching keys.
func (r *Client) GetKeys(pattern string) []string {
	val, err := r.client.Keys(pattern).Result()

	if err != nil {
		l.ErrorLog(err, "There was an error while getting keys from redis. Key pattern was {{.pattern}}", map[string]interface{}{
			"pattern": pattern,
		})
		return []string{}
	}

	l.DebugLog("There has been {{.count}} keys for pattern '{{.pattern}}'", map[string]interface{}{
		"pattern":      pattern,
		"redis_result": val,
		"count":        len(val),
	})
	return val

}

// DeleteKey calls the redis del function and returns the result.
// If there was an error zero is returned.
func (r *Client) DeleteKey(key string) int64 {
	val, err := r.client.Del(key).Result()

	if err != nil {
		l.ErrorLog(err, "There was an error while deleting {{.redis_key}} from redis.", map[string]interface{}{
			"redis_key": key,
		})
		return 0
	}

	l.DebugLog("The key {{.redis_key}} has been deleted", map[string]interface{}{
		"redis_key":    key,
		"redis_result": val,
	})
	return val
}

func (r *Client) increaseStatisticCount(name string, t string) int64 {
	ts := int(time.Now().Unix())
	timestampDay := ts - int(math.Mod(float64(ts), float64(24*60*60)))
	val, err := r.client.HIncrBy("stats:"+name+":"+fmt.Sprint(timestampDay), t, int64(1)).Result()

	if err != nil {
		l.ErrorLog(err, "There was an error while increasing stats [{{.stat_type}}] for name '{{.redis_key}}'.", map[string]interface{}{
			"redis_key": name,
			"stat_type": t,
		})
		return int64(0)
	}

	l.DebugLog("Stats counter [{{.stat_type}}] for key '{{.redis_key}}' is now at {{.redis_val}}", map[string]interface{}{
		"redis_key":    name,
		"redis_result": val,
		"stat_type":    t,
	})
	return val
}

// IncreaseStatisticCountMail is used to count mails per name without expiring the count.
// This is used to check if a rules has any hits.
func (r *Client) IncreaseStatisticCountMail(name string) int64 {
	return r.increaseStatisticCount(name, "mail")
}

// IncreaseStatisticCountWarning is used to count warning alerts per name without expiring the count.
// This is used to check if a rules has any hits.
func (r *Client) IncreaseStatisticCountWarning(name string) int64 {
	return r.increaseStatisticCount(name, "warning")
}

// IncreaseStatisticCountCritical is used to count critical alerts per name without expiring the count.
// This is used to check if a rules has any hits.
func (r *Client) IncreaseStatisticCountCritical(name string) int64 {
	return r.increaseStatisticCount(name, "critical")
}

// GetStatisticCount returns the actual number of hits for one name.
func (r *Client) GetStatisticCount(name string, timestamp int) Stats {
	stats := Stats{Name: name, Mail: 0, Warning: 0, Critical: 0}
	timestampDay := timestamp - int(math.Mod(float64(timestamp), float64(24*60*60)))
	val, err := r.client.HMGet("stats:"+name+":"+fmt.Sprint(timestampDay), "mail", "warning", "critical").Result()

	if err != nil {
		l.ErrorLog(err, "There was an error while getting stats for name '{{.redis_key}}'.", map[string]interface{}{
			"redis_key": name,
			"stats":     stats,
		})
		return stats
	}

	stats.Mail = checkVal(val[0], name)
	stats.Warning = checkVal(val[1], name)
	stats.Critical = checkVal(val[2], name)

	l.DebugLog("The stats counter for name '{{.redis_key}}' is at {{.redis_result}}", map[string]interface{}{
		"redis_key":    name,
		"redis_result": val,
		"stats":        stats,
	})
	return stats
}

func checkVal(val interface{}, name string) int64 {
	if val != nil {
		result, err := strconv.ParseInt(fmt.Sprint(val), 10, 64)
		if err != nil {
			l.ErrorLog(err, "There was an error while parsing stats counter value [{{.stats_type}}] for key '{{.name}}' from redis. value was {{.redis_result}}", map[string]interface{}{
				"stats_type":   "critical",
				"redis_result": val,
				"name":         name,
			})
		} else {
			return result
		}
	} else {
		return 0
	}
	return 0
}
