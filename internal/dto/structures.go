package dto

import (
	"time"

	sqlc "go.mod/internal/sqlc/generate"
	"google.golang.org/api/forms/v1"
)

type NewJobData struct {
	CompanyName string
	CompanyEmail string
	JobLocation string
	JobTitle string
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
	Date time.Time `form:"Date" time_format:"2006-01-02"`
	Time time.Time `form:"Time" time_format:"15:04"`
	DateTime time.Time
	FormattedTime string
	FormattedDate string
	Type string
	Location string
	Notes string
	StudentName string
	StudentEmail string
	JobTitle string
	CompanyName string
}

type Offer struct {
	ApplicationId int64 `form:"ApplicationId"`
}

type CancelInterview struct {
	StudentName string
	StudentEmail string
	JobTitle string
	CompanyName string
	Date string
	Time string
	RepresentativeEmail string
	RepresentativeName string
}

type AllUpcomingData struct {
	InterviewsData []sqlc.GetInterviewsForUserIDRow
	TestsData []sqlc.GetTestsForUserIDRow
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
}
