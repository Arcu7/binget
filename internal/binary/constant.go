package binary

const (
	GithubHost   = "github.com"
	GithubAPIURL = "https://api.github.com/repos/%s/%s/releases/latest"
)

type FileType string

const (
	FileTypeZip   FileType = ".zip"
	FileTypeTarGz FileType = ".tar.gz"
)
