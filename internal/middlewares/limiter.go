package middlewares

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/config"
)

// type Prolonged struct {
// 	LastRecordTime int64
// 	ReqCounter int64
// }

// var (
// 	prolongedMap = make(map[string]*Prolonged)
// )

var (
	rScript = `
		local key = KEYS[1]
		local bucket_size = tonumber(ARGV[1])
		local expiry = tonumber(ARGV[2])

		local current = redis.call("GET",key)

		if not current then 
			redis.call("SET", key, bucket_size - 1, "PX", expiry)
			return bucket_size - 1
		else 
			current = tonumber(current)
		end

		if current <= 0 then 
			return -1
		else 
			redis.call("DECR", key)
			return current - 1
		end
	` 
)


func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ip := ctx.ClientIP()
		nparse := net.ParseIP(ip)
		if ip == "" || nparse == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

		result, err := redisClient.Eval(ctx, rScript, []string{ip}, config.RateLimiterBucketSize, config.RateLimiterExpiry).Int()
		
		if err != nil {
			ctx.Set("critical", "Rate Limiter failed to execute Lua script : " + err.Error())
			fmt.Println(err)
			return
		}
		if result == -1 {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>


		// data, exists := prolongedMap[ip]
		// if !exists {
		// 	prolongedMap[ip] = &Prolonged{
		// 		LastRecordTime: time.Now().Unix(),
		// 		ReqCounter: 1,
		// 	}
		// } else {
		// 	data.ReqCounter++

		// 	secSince := time.Now().Unix() - data.LastRecordTime 
		// 	reqSince := data.ReqCounter

		// 	if secSince > 30 {
		// 		data.LastRecordTime = time.Now().Unix()
		// 		data.ReqCounter = 1
		// 	} 
		// 	if reqSince > config.RequestWindowCounter {
		// 		ctx.AbortWithStatus(http.StatusForbidden)
		// 		return
		// 	}
			
			
		// }


		ctx.Next()
	}
}