package middlewares

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/config"
)


type Prolonged struct {
	LastRecordTime int64
	ReqCounter int64
}

var (
	prolongedMap = make(map[string]*Prolonged)
)







type RequestRate struct {
	FirstRequestAt int64
	NumberOfReqsSince int64
	StrikerCounter int8
	LastExceededAt int64
}
var (
	rwmu sync.RWMutex
	reqRateMap = make(map[string]*RequestRate)
)
// TODO: need to clean up the map bro periodically
func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ip := ctx.ClientIP()
		nparse := net.ParseIP(ip)
		if ip == "" || nparse == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		// TODO: maybe replace the get-set with getbit-setbit ??
		// basic ip limiting, uses token buckets to allow client-side JS to load data

		// TODO: this is stupidity, there's no locks no atomicity, nothing, it works for sequential hits from a single worker
		// the moment i increase the workers, i can get way more hits through there, like 25 

		val, err := redisClient.Get(ctx, ip).Result()
		if err != nil {
			if err == redis.Nil {
				err := redisClient.Set(ctx, ip, config.RateLimiterBucketSize, time.Duration(config.RateLimiterExpiry) * time.Millisecond).Err()
				if err != nil {
					ctx.Set("critical", "Rate Limiter : Failed to Set keys in redis : " + err.Error())
					return
				}
			} else {
				ctx.Set("critical", "Rate Limiter : Failed to get keys : " + err.Error())
				return
			}	
		} else {

			value, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				ctx.Set("critical", "Failed to parse key - value from string to int64 : " + err.Error())
				return
			}

			fmt.Println(value)

			if value <= 0 {
				ctx.Header("Retry-After", "60") // TODO: this does not work properly
				ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "Too many requests. Try again later.",
				})
				return
			} else {
				err = redisClient.DecrBy(ctx, ip, 1).Err()
				if err != nil {
					ctx.Set("critical", "Failed to decrement key - value (current bucket size) : " + err.Error())
					return
				}
			}
		}
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>


		data, exists := prolongedMap[ip]
		if !exists {
			prolongedMap[ip] = &Prolonged{
				LastRecordTime: time.Now().Unix(),
				ReqCounter: 1,
			}
		} else {
			data.ReqCounter++

			secSince := time.Now().Unix() - data.LastRecordTime 
			reqSince := data.ReqCounter

			if secSince > 30 {
				data.LastRecordTime = time.Now().Unix()
				data.ReqCounter = 1
			} 
			if reqSince > config.RequestWindowCounter {
				ctx.AbortWithStatus(http.StatusForbidden)
				return
			}
			
			
		}















		ctx.Next()
	}
}