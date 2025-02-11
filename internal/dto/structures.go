package dto

import (
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	sqlc "go.mod/internal/sqlc/generate"
	"google.golang.org/api/forms/v1"
)

type NewJobData struct {
	JobId int64
	JobTitle string
	JobLocation string
	JobDescription string
	JobType string
	JobSalary string
	SkillsRequired string
	JobPosition string
	Extras map[string]interface{}
}

type AllJobs struct {
    ID         int    		`json:"id"`
	Title       string		`json:"title"`
    Location    string	`json:"location"`
    Type        string	`json:"type"`
    Salary      string	`json:"salary"`
    Position    string	`json:"position"`
    Skills      []string	`json:"skills"`
    Extras      []byte	`json:"extras"`
    CompanyName string	`json:"company_name"`
}

type ExtraInfoCompany struct {
	Token string
	CompanyName string
	CompanyEmail string
	CompanyDescription string
	CompanyAddress string
	CompanyWebsite string
	IndustryType string

	RepresentativeName string
	RepresentativeContact string
	RepresentativeEmail string

	Extras map[string]interface{}
}

type ExtraInfoStudent struct {
	Token string
	StudentName string
	CollegeRollNumber string
	DateOfBirth time.Time `form:"DateOfBirth" time_format:"2006-01-02"`  
	Gender string
	Course string
	Department string
	YearOfStudy string
	ResumeUrl string       
	ResultUrl string		
	CGPA float64
	ContactNumber string
	StudentEmail string
	Address string
	Skills string
	Extras map[string]interface{}
}

type NewInterview struct {
	ApplicationId int64
	UserId int64
	DateTime time.Time `form:"DateTime" time_format:"2006-01-02T15:04"`
	Type string
	Location string
	Notes string
	StudentName string
	StudentEmail string
	JobTitle string
	CompanyName string
	DT string
}

type UpdateInterview struct {
	InterviewID int64
	DateTime time.Time `form:"DateTime" time_format:"2006-01-02T15:04"`
	Type string
	Location string
	Notes string

	StudentName string
	JobTitle string
	CompanyName string
	DT string
}

type Offer struct {
	ApplicationId int64 `form:"ApplicationId"`
}

type CancelInterview struct {
	StudentName string
	StudentEmail string
	JobTitle string
	CompanyName string
	DateTime string
	RepresentativeEmail string
	RepresentativeName string
}

type Upcoming struct {
	Data any
}

type Completed struct {
	Data any
}

type NewTestPost struct {
	Name string
	Description string
	Duration int64
	QuestionCount int64
	EndDateTime time.Time `form:"EndDateTime" time_format:"2006-01-02T15:04"`
	BindedJobId int64
	Type string
	UploadMethod string
	Threshold int64

	// TODO: this is kinda useless, remove it !
	FormattedEndDate string
	FormattedEndTime string 
	JobTitle string
	CompanyName string
	Link string
}

type NewTestGForms struct {
	ResponderLink string
	Notes string
}

type TestQuestion struct {
	Item *forms.Item
	PrevId string
	NextId string
	TTL time.Duration
}

type TestResponse struct {
	ItemID string
	Response []string
	TimeTaken int64
}
// TODO: replace this later with the 'NewTestPost' struct
type UpdateTest struct {
	TestID int64
	Threshold int64
}

type Token struct {
	Issuer string
	Subject string
	ExpiresAt int64
	IssuedAt int64
	Role int64
	ID int64
	Email string	

	Version string
}

type JWTTokens struct {
	JWTAccess string
	JWTRefresh string
}

type StudentProfileData struct {
	OverData *sqlc.ApplicationsStatusCountsRow
	UsersData *sqlc.UsersTableDataRow
	ProData *sqlc.StudentProfileDataRow
	AppsHistory *[]sqlc.ApplicationHistoryRow
	IntsHistory *[]sqlc.InterviewHistoryRow
	TestsHistory *[]sqlc.TestHistoryRow
	SankeyChrt *charts.Sankey
}

type CompanyProfileData struct {
	OverData *sqlc.ApplicantsCountRow
	UsersData *sqlc.UsersTableDataRow
	ProData *sqlc.CompanyProfileDataRow
	// AppsHistory *[]sqlc.ApplicationHistoryRow
	// IntsHistory *[]sqlc.InterviewHistoryRow
	// TestsHistory *[]sqlc.TestHistoryRow
	SankeyChrt []*charts.Sankey
}

type UpdateStudentDetails struct {
	Course string
	Department string
	YearOfStudy string
	CGPA float64
	ContactNumber string
	Address string
	Skills string
}

type UpdateCompanyDetails struct {
	CompanyName string
	CompanyDescription string
	CompanyAddress string
	IndustryType string
	CompanyWebsite string
	RepresentativeName string
	RepresentativeEmail string
	RepresentativeContact string
}

type Report struct {
	UserId int64
	Message string
	ReportedAt time.Time
	IpAddress string
}

type CumulativeChartsData struct {
	Xaxis []string
	Yaxis []int64

	PassCount int64
	FailCount int64
}

type IndividualChartsData struct {
	FunnelDimensions []string
	FunnelValues []int64

	RadarNames []*opts.Indicator
	RadarValues []float32
}


type NotificationData struct {
	Title string
	Description string
	TimeStamp int64
}






type PerformanceData struct {
	Latency float64
}

type LoggerData struct {
	StartTime time.Time
	ClientIP string
	Method string
	Path string
	StatusCode int
	InternalError string
	Latency time.Duration
}

type ErrorData struct {
	Debug string // found in ctx.Get("debug")
	Info string // found in ctx.Get("info")
	Warn string // found in ctx.Get("warn")
	Error string // found in ctx.Get("error")
	Critical string // found in ctx.Get("critical")
	Fatal string // found in ctx.Get("fatal")

	LogData *LoggerData
}