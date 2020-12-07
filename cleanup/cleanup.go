package cleanup

import (
	"strconv"
	"strings"
	"time"

	"niecke-it.de/veloci-meter/config"
	l "niecke-it.de/veloci-meter/logging"
	"niecke-it.de/veloci-meter/rdb"
	"niecke-it.de/veloci-meter/rules"
)

// CleanUp removes data for global rules which are older than one day.
func CleanUp(conf *config.Config) {
	l.DebugLog("Running clean up job.", nil)
	timestamp := int(time.Now().Unix())
	deletedKey := 0
	l.DebugLog("Connecting to redis...", map[string]interface{}{"redis_uri": conf.Redis.URI})
	r := rdb.NewClient(&conf.Redis)

	for index, val := range rules.GlobalPatterns {
		l.DebugLog("Checking {{.index}} keys...", map[string]interface{}{"index": index})
		keys := r.GetKeys(val + "*")
		for _, key := range keys {
			ts, err := strconv.Atoi(strings.Replace(key, val, "", -1))
			if err != nil {
				l.ErrorLog(err, "There was an error converting {{.data}} to int.", map[string]interface{}{"data": key})
			} else if timestamp-ts > 86400 {
				// if the key is older than 24 hours -> delete it

				redisReturn := r.DeleteKey(key)
				l.InfoLog("Redis return for deleting {{.redis_key}} was {{.redis_result}}", map[string]interface{}{"redis_key": key, "redis_result": redisReturn})
				deletedKey++
			}
		}
	}
	end := int(time.Now().Unix())
	duration := end - timestamp
	l.InfoLog("Cleanup job is done. Deleted {{.redis_key}} keys from redis in {{.duration}} seconds.", map[string]interface{}{"redis_key": deletedKey, "duration": duration})
}
