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

const (
	ResumeFileSizeLimit = 300000 // Bytes
	ResultFileSizeLimit = 300000 // Bytes
	ProfilePicFileSizeLimit = 300000 // Bytes
)

var (
	ResumeFileContentTypes = []string{"application/pdf"} // string slice of extesions permissible 
	ResultFileContentTypes = []string{"application/pdf"} // string slice of extesions permissible 
	ProfilePicFileContentTypes = []string{"image/jpeg", "image/jpg", "image/png"} // string slice of extesions permissible 
)

var (
	FileSizeForContentType = map[string]int64{
		"application/pdf": 300000,

		"image/jpeg": 300000, 
		"image/jpg": 300000, 
		"image/png": 300000,
	}
)