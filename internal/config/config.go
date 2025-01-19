package config

const (
	JWTAccessExpiration = 60 // mins // reduce later 
	JWTRefreshExpiration = 168 // hours // 7 days 
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