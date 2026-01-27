package binary

import "fmt"

const (
	GithubHost   = "github.com"
	GithubAPIURL = "https://api.github.com/repos/%s/%s/releases/latest"
)

type FileType string

const (
	FileTypeZIP   FileType = ".zip"
	FileTypeTarGz FileType = ".tar.gz"
)

func parseFileType(s string) (FileType, error) {
	ft := FileType(s)
	switch ft {
	case FileTypeZIP, FileTypeTarGz:
		return ft, nil
	default:
		return "", fmt.Errorf("invalid file type: %s", s)
	}
}
