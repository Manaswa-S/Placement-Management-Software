package dto

import (
	"time"
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
	JobId int64
	StudentId int64
	UserId int64
	Date time.Time `form:"Date" time_format:"2006-01-02"`
	Time time.Time `form:"Time" time_format:"15:04"`
	Type string
	Notes string
}