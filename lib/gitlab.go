package stew

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/marwanhawari/stew/constants"
)

// GitlabProject contains information about the Gitlab project including Gitlab releases
type GitlabProject struct {
	Owner    string
	Repo     string
	Releases GitlabAPIResponse
}

// GitlabAPIResponse is the response from the Gitlab releases API
type GitlabAPIResponse []GitlabRelease

// GitlabRelease contains information about a Gitlab release, including the associated assets
type GitlabRelease struct {
	TagName string        `json:"tag_name"`
	ID      int           `json:"id"`
	Assets  []GitlabAsset `json:"assets"`
}

// GitlabAsset contains information about a specific Gitlab asset
type GitlabAsset struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
}

func readGitlabJSON(host, owner, repo, jsonString string) (GitlabAPIResponse, error) {
	var gtProject GitlabAPIResponse
	err := json.Unmarshal([]byte(jsonString), &gtProject)
	if err != nil {
		return GitlabAPIResponse{}, err
	}
	return gtProject, nil
}

func getGitlabJSON(host, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://%s/api/v1/repos/%v/%v/releases?per_page=100", host, owner, repo)

	response, err := getHTTPResponseBody(url, "gitlab")
	if err != nil {
		return "", err
	}

	return response, nil
}

// NewGitlabProject creates a new instance of the GitlabProject struct
func NewGitlabProject(host, owner, repo string) (GitlabProject, error) {
	gtJSON, err := getGitlabJSON(host, owner, repo)
	if err != nil {
		return GitlabProject{}, err
	}

	gtAPIResponse, err := readGitlabJSON(host, owner, repo, gtJSON)
	if err != nil {
		return GitlabProject{}, err
	}

	ghProject := GitlabProject{Owner: owner, Repo: repo, Releases: gtAPIResponse}

	return ghProject, nil
}

// GetGitlabReleasesTags gets a string slice of the releases for a GitlabProject
func GetGitlabReleasesTags(ghProject GitlabProject) ([]string, error) {
	releasesTags := []string{}

	for _, release := range ghProject.Releases {
		releasesTags = append(releasesTags, release.TagName)
	}

	err := releasesFound(releasesTags, ghProject.Owner, ghProject.Repo)
	if err != nil {
		return []string{}, err
	}

	return releasesTags, nil
}

// GetGitlabReleasesAssets gets a string slice of the assets for a GitlabRelease
func GetGitlabReleasesAssets(ghProject GitlabProject, tag string) ([]string, error) {
	releaseAssets := []string{}

	for _, release := range ghProject.Releases {
		if release.TagName == tag {
			for _, asset := range release.Assets {
				releaseAssets = append(releaseAssets, asset.Name)
			}
		}
	}

	err := assetsFound(releaseAssets, tag)
	if err != nil {
		return []string{}, err
	}

	return releaseAssets, nil
}

func getGitlabSearchJSON(host, searchQuery string) (string, error) {
	url := fmt.Sprintf("https://%s/api/v1/repos/search?q=%v", host, searchQuery)

	response, err := getHTTPResponseBody(url, "gitlab")
	if err != nil {
		return "", err
	}

	return response, nil
}

func readGitlabSearchJSON(jsonString string) (RepoSearch, error) {
	var gtSearch RepoSearch
	err := json.Unmarshal([]byte(jsonString), &gtSearch)
	if err != nil {
		return RepoSearch{}, err
	}
	return gtSearch, nil
}

// NewGitlabSearch creates a new instance of the GitlabSearch struct
func NewGitlabSearch(host, searchQuery string) (RepoSearch, error) {
	gtJSON, err := getGitlabSearchJSON(host, searchQuery)
	if err != nil {
		return RepoSearch{}, err
	}

	gtSearch, err := readGitlabSearchJSON(gtJSON)
	if err != nil {
		return RepoSearch{}, err
	}

	gtSearch.SearchQuery = searchQuery

	return gtSearch, nil
}

// FormatSearchResults formats the Gitlab search results for the terminal UI
func FormatGitlabSearchResults(ghSearch RepoSearch) []string {
	var formattedSearchResults []string
	for _, searchResult := range ghSearch.Items {
		formatted := fmt.Sprintf("%v [⭐️%v] %v", searchResult.FullName, searchResult.Stars, searchResult.Description)
		formattedSearchResults = append(formattedSearchResults, formatted)
	}

	return formattedSearchResults
}

// ValidateGitetatSearchQuery makes sure the Gitlab search query is valid
func ValidateGitlabSearchQuery(searchQuery string) error {
	reSearch, err := regexp.Compile(constants.RegexGitlabSearch)
	if err != nil {
		return err
	}

	if !reSearch.MatchString(searchQuery) {
		return InvalidGitlabSearchQueryError{SearchQuery: searchQuery}
	}

	return nil
}
