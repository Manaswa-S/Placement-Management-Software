package dto

type NewJobData struct {
	CompanyName string
	CompanyEmail string
	CompanyLocation string
	JobTitle string
	JobDescription string
	JobType string
	JobSalary string
	SkillsRequired string
	JobPosition string
	Extras map[string]interface{}
}


type ExtraInfoCompany struct {
	Token string
	CompanyName string
	RepresentativeName string
	RepresentativeContact string
	Extras map[string]interface{}
}

type ExtraInfoStudent struct {
	Token string
	StudentName string
	Gender string
	CollegeRollNumber string
	Department string
	Extras map[string]interface{}
}