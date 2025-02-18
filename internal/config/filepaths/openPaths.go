package filepaths


// :Path at last represents the complete path to the file from the root directory
// :Name at last represents only the file name with its extension
type OpenPaths struct {
	DiscussionsPagePath string
}


// Direct (the file, not basepaths) paths to all Dashboards
func LoadOpenPaths() OpenPaths {
	return OpenPaths{
		DiscussionsPagePath: "./template/open/discussions.html",
	}
}
