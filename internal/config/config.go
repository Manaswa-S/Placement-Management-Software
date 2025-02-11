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
	Paths = filepaths.LoadFilePaths()
)


