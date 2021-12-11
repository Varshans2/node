package ui

import (
	"fmt"
	"github.com/mysteriumnetwork/node/requests"
	"github.com/pkg/errors"
	"time"
)

const (
	nodeUIAssetName = "dist.tar.gz"
	apiURI          = "https://api.github.com"
	nodeUIPath      = "repos/mysteriumnetwork/dvpn-web"
)

type github struct {
	http *requests.HTTPClient
}

func newGithub(httpClient *requests.HTTPClient) *github {
	return &github{http: httpClient}
}

func (g *github) nodeUIReleases() ([]GitHubRelease, error) {
	req, err := requests.NewGetRequest(apiURI, fmt.Sprintf("%s/releases", nodeUIPath), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create NodeUI releases fetch request: %w", err)
	}

	var releases []GitHubRelease
	err = g.http.DoRequestAndParseResponse(req, &releases)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NodeUI releases: %w", err)
	}
	return releases, nil
}

func (g *github) nodeUIDownloadURL(versionName string) (string, error) {
	releases, err := g.nodeUIReleases()
	if err != nil {
		return "", fmt.Errorf("failed to fetch NodeUI releases: %w", err)
	}

	r, ok := findRelease(releases, versionName)
	if !ok {
		return "", errors.New(fmt.Sprintf("could not find release named %s", versionName))
	}

	req, err := requests.NewGetRequest(apiURI, fmt.Sprintf("%s/releases/%d/assets", nodeUIPath, r.Id), nil)
	if err != nil {
		return "", fmt.Errorf(fmt.Sprintf("failed to create NodeUI assets ID %d, versionName %s", r.Id, r.Name)+": %w", err)
	}

	var assets []GithubAsset
	err = g.http.DoRequestAndParseResponse(req, &assets)
	if err != nil {
		return "", fmt.Errorf("failed to fetch NodeUI releases: %w", err)
	}

	uiDist, ok := findNodeUIDist(assets)
	if !ok {
		return "", errors.New(fmt.Sprintf("could not find nodeUI dist asset for release ID %d version %s", r.Id, versionName))
	}
	return uiDist.BrowserDownloadUrl, nil
}

func findNodeUIDist(assets []GithubAsset) (GithubAsset, bool) {
	for _, ass := range assets {
		if ass.Name == nodeUIAssetName {
			return ass, true
		}
	}
	return GithubAsset{}, false
}

func findRelease(releases []GitHubRelease, versionName string) (GitHubRelease, bool) {
	for _, release := range releases {
		if release.Name == versionName {
			return release, true
		}
	}
	return GitHubRelease{}, false
}

type GitHubRelease struct {
	Url       string `json:"url"`
	AssetsUrl string `json:"assets_url"`
	UploadUrl string `json:"upload_url"`
	HtmlUrl   string `json:"html_url"`
	Id        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		Id                int    `json:"id"`
		NodeId            string `json:"node_id"`
		AvatarUrl         string `json:"avatar_url"`
		GravatarId        string `json:"gravatar_id"`
		Url               string `json:"url"`
		HtmlUrl           string `json:"html_url"`
		FollowersUrl      string `json:"followers_url"`
		FollowingUrl      string `json:"following_url"`
		GistsUrl          string `json:"gists_url"`
		StarredUrl        string `json:"starred_url"`
		SubscriptionsUrl  string `json:"subscriptions_url"`
		OrganizationsUrl  string `json:"organizations_url"`
		ReposUrl          string `json:"repos_url"`
		EventsUrl         string `json:"events_url"`
		ReceivedEventsUrl string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeId          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		Url      string `json:"url"`
		Id       int    `json:"id"`
		NodeId   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			Id                int    `json:"id"`
			NodeId            string `json:"node_id"`
			AvatarUrl         string `json:"avatar_url"`
			GravatarId        string `json:"gravatar_id"`
			Url               string `json:"url"`
			HtmlUrl           string `json:"html_url"`
			FollowersUrl      string `json:"followers_url"`
			FollowingUrl      string `json:"following_url"`
			GistsUrl          string `json:"gists_url"`
			StarredUrl        string `json:"starred_url"`
			SubscriptionsUrl  string `json:"subscriptions_url"`
			OrganizationsUrl  string `json:"organizations_url"`
			ReposUrl          string `json:"repos_url"`
			EventsUrl         string `json:"events_url"`
			ReceivedEventsUrl string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadUrl string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballUrl string `json:"tarball_url"`
	ZipballUrl string `json:"zipball_url"`
	Body       string `json:"body"`
}

type GithubAsset struct {
	Url      string `json:"url"`
	Id       int    `json:"id"`
	NodeId   string `json:"node_id"`
	Name     string `json:"name"`
	Label    string `json:"label"`
	Uploader struct {
		Login             string `json:"login"`
		Id                int    `json:"id"`
		NodeId            string `json:"node_id"`
		AvatarUrl         string `json:"avatar_url"`
		GravatarId        string `json:"gravatar_id"`
		Url               string `json:"url"`
		HtmlUrl           string `json:"html_url"`
		FollowersUrl      string `json:"followers_url"`
		FollowingUrl      string `json:"following_url"`
		GistsUrl          string `json:"gists_url"`
		StarredUrl        string `json:"starred_url"`
		SubscriptionsUrl  string `json:"subscriptions_url"`
		OrganizationsUrl  string `json:"organizations_url"`
		ReposUrl          string `json:"repos_url"`
		EventsUrl         string `json:"events_url"`
		ReceivedEventsUrl string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadUrl string    `json:"browser_download_url"`
}
