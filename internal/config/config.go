package config

import "go.mod/internal/config/filepaths"

// unit // conversion

const (
	JWTAccessExpiration = 3600 // seconds //  
	JWTRefreshExpiration = 604800 // seconds // 7 days // 604800 seconds
)

const (
	TestResultPollerTimeout = 900 // seconds // 15 mins
)

const (
	SignupConfirmLinkTokenExpiration = 15 // mins
	ResetLinkTokenExpiration = 15 // mins
)

const (
	TempFileStorage = "./temp"
)

var (
	FileSizeForContentType = map[string]int64{
		"application/pdf": 300000, // bytes

		"image/jpeg": 300000, 
		"image/jpg": 300000, 
		"image/png": 300000,
	}
)

const (
	PingTesterIP = "8.8.8.8"
)

const (
	TestResultTaskQueueBufferCapacity = 100
	TestResultFailedQueueBufferCapacity = 10

	NoOfTestResultTaskWorkers = 1
	NoOfTestResultFailedWorkers = 1
)

const (
	// 0 : infinite blocking
	// x : waits for x milliseconds to return
	NotificationsXReadBlock = 1 
	// number of notifications to read in a batch
	NotificationsXReadCount = 10 
)



var (
	Paths = filepaths.LoadFilePaths() // TODO: remove this
	CompPaths = filepaths.LoadCompanyPaths()
	OpenPaths = filepaths.LoadOpenPaths()
)

const (
	DiscussionPageLimit = 10
	DiscussionPageAutoRefreshRate = 15000; // in milliseconds, the page auto refreshes in 
)


const (
	// This rate limit is a very good combination. (3/1000)
	// It allows for continuous reloading of pages one after another,
	// but as soon as you start spamming ctrl+R it blocks.
	// This is on a very fast 5G network and will get better for slower networks.
	// {RateLimiterBucketSize} requests per {RateLimiterExpiry} milliseconds.
	RateLimiterBucketSize = 4 // in int only
	RateLimiterExpiry = 1000 // in milliseconds only

	// Prolonged Request Rate Tracker (Sustained High Rate Over Time)
	// Ban an IP for RequestRateTempBanExpiry if the average request rate (avg(number of reqs / time window)) exceeeds RequestRateLimit.
	// We keep this limit slightly lower than the RateLimiter limit.
	// This ensures that even if a user stays just below the short-term limit,
	// they cannot sustain a high request rate indefinitely.
	
	// in milliseconds only, this is considered as a factor, 
	// the actual ban period depends on the number of times the threshold was crossed, StrikeCounter
	RequestRateTempBanExpiry = 900000 
	RequestRateLimit float32 = 3.0 // in float32 only	
	RequestRateRefreshAfter = 30 // in seconds only
	RequestRateStrikeCounterLimit = 20 // in int8 only // if greater than 

	RequestWindowCounter = 100
)

