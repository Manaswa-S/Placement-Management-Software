package filepaths


// :Path at last represents the complete path to the file from the root directory
// :Name at last represents only the file name with its extension


type StudentPaths struct {
	DashboardPath string
	JobListingsTemplatePath string
	ApplicationsTemplatePath string
	FeedbacksTemplatePath string

}


func LoadStudentPaths() StudentPaths {
	return StudentPaths{
		DashboardPath: "./template/dashboard/studentdashboard.html",
		JobListingsTemplatePath: "./template/student/alljobslist.html",
		ApplicationsTemplatePath: "./template/student/myapplications.html",
		FeedbacksTemplatePath: "./template/student/feedbacks.html",
	}
}