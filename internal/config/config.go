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