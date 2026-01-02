package binary

type Request struct {
	RepositoryURL string
	AuthToken     string
	OS            string
	Arch          string
	PathEnv       string
}

type GithubReleaseResponse struct {
	Assets []GithubReleaseAssetsResponse `json:"assets"`
}

type GithubReleaseAssetsResponse struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}
