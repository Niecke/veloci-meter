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
	// using a random int32 as part of the redis key
	random_part, _ := rand.Int(rand.Reader, big.NewInt(2147483647))
	r.client.Set(sha1_hash+":"+fmt.Sprint(random_part), 1, time.Duration(duration)*time.Second)
	l.Debugln("Stored " + sha1_hash + ":" + fmt.Sprint(random_part) + " for " + fmt.Sprint(time.Duration(duration)*time.Second))
}

func (r *RDBClient) CountMail(pattern string) int64 {
	sha1_hash := buildHash(pattern)
	v, err := r.client.Eval("return #redis.pcall('keys', '"+sha1_hash+":*')", nil).Result()
	if err != nil {
		l.Fatal(err)
	}
	return v.(int64)
}

func calculateGlobalKey(timestamp int, timeframe int) string {
	remainder := math.Mod(float64(timestamp), float64(timeframe*60))
	key_part := timestamp - int(remainder)
	return "global:" + fmt.Sprint(timeframe) + ":" + fmt.Sprint(key_part)
}

// timefram in minutes
func (r *RDBClient) IncreaseGlobalCounter(timeframe int) {
	timestamp := int(time.Now().Unix())
	redis_key := calculateGlobalKey(timestamp, timeframe)
	err := r.client.Incr(redis_key)
	if err != nil {
		l.Debugf("[%v] Redis Command executed: [%v]", err, redis_key)
	}
}

func (r *RDBClient) GetGlobalCounter(timeframe int) int {
	timestamp := int(time.Now().Unix())
	redis_key := calculateGlobalKey(timestamp, timeframe)
	val, err := r.client.Get(redis_key).Result()

	if err != nil {
		if err == redis.Nil {
			return 0
		}
		l.Errorf("[%v] There was an error while getting global counter from redis. Redis key was %v", err, redis_key)
	}

	c, err := strconv.Atoi(val)
	if err != nil {
		l.Errorf("[%v] There parsing global counter value from redis. value was %v", err, val)
	}

	return c
}

func (r *RDBClient) GetKeys(pattern string) []string {
	val, err := r.client.Keys(pattern).Result()

	if err != nil {
		l.Errorf("[%v] There was an error while getting keys from redis. Key pattern was %v", err, pattern)
	}

	return val
}

func (r *RDBClient) DeleteKey(key string) int64 {
	val, err := r.client.Del(key).Result()

	if err != nil {
		l.Errorf("[%v] There was an error while deleting %v from redis.", err, key)
	}

	return val
}
