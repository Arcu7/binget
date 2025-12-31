package binary

type Request struct {
	RepositoryURL string
	AuthToken     string
}

type GithubReleaseResponse struct {
	Assets []GithubReleaseAssetsResponse `json:"assets"`
}

type GithubReleaseAssetsResponse struct {
	BrowserDownloadURL string `json:"browser_download_url"`
}
