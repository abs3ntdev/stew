package stew

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/marwanhawari/stew/constants"
)

// GiteaProject contains information about the Gitea project including Gitea releases
type GiteaProject struct {
	Owner    string
	Repo     string
	Releases GiteaAPIResponse
}

// GiteaAPIResponse is the response from the Gitea releases API
type GiteaAPIResponse []GiteaRelease

// GiteaRelease contains information about a Gitea release, including the associated assets
type GiteaRelease struct {
	TagName string       `json:"tag_name"`
	ID      int          `json:"id"`
	Assets  []GiteaAsset `json:"assets"`
}

// GiteaAsset contains information about a specific Gitea asset
type GiteaAsset struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
}

func readGiteaJSON(host, owner, repo, jsonString string) (GiteaAPIResponse, error) {
	var gtProject GiteaAPIResponse
	err := json.Unmarshal([]byte(jsonString), &gtProject)
	if err != nil {
		return GiteaAPIResponse{}, err
	}
	return gtProject, nil
}

func getGiteaJSON(host, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://%s/api/v1/repos/%v/%v/releases?per_page=100", host, owner, repo)

	response, err := getHTTPResponseBody(url, "gitea")
	if err != nil {
		return "", err
	}

	return response, nil
}

// NewGiteaProject creates a new instance of the GiteaProject struct
func NewGiteaProject(host, owner, repo string) (GiteaProject, error) {
	gtJSON, err := getGiteaJSON(host, owner, repo)
	if err != nil {
		return GiteaProject{}, err
	}

	gtAPIResponse, err := readGiteaJSON(host, owner, repo, gtJSON)
	if err != nil {
		return GiteaProject{}, err
	}

	ghProject := GiteaProject{Owner: owner, Repo: repo, Releases: gtAPIResponse}

	return ghProject, nil
}

// GetGiteaReleasesTags gets a string slice of the releases for a GiteaProject
func GetGiteaReleasesTags(ghProject GiteaProject) ([]string, error) {
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

// GetGiteaReleasesAssets gets a string slice of the assets for a GiteaRelease
func GetGiteaReleasesAssets(ghProject GiteaProject, tag string) ([]string, error) {
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

func getGiteaSearchJSON(host, searchQuery string) (string, error) {
	url := fmt.Sprintf("https://%s/api/v1/repos/search?q=%v", host, searchQuery)

	response, err := getHTTPResponseBody(url, "gitea")
	if err != nil {
		return "", err
	}

	return response, nil
}

func readGiteaSearchJSON(jsonString string) (RepoSearch, error) {
	var gtSearch RepoSearch
	err := json.Unmarshal([]byte(jsonString), &gtSearch)
	if err != nil {
		return RepoSearch{}, err
	}
	return gtSearch, nil
}

// NewGiteaSearch creates a new instance of the GiteaSearch struct
func NewGiteaSearch(host, searchQuery string) (RepoSearch, error) {
	gtJSON, err := getGiteaSearchJSON(host, searchQuery)
	if err != nil {
		return RepoSearch{}, err
	}

	gtSearch, err := readGiteaSearchJSON(gtJSON)
	if err != nil {
		return RepoSearch{}, err
	}

	gtSearch.SearchQuery = searchQuery

	return gtSearch, nil
}

// FormatSearchResults formats the Gitea search results for the terminal UI
func FormatGiteaSearchResults(ghSearch RepoSearch) []string {
	var formattedSearchResults []string
	for _, searchResult := range ghSearch.Items {
		formatted := fmt.Sprintf("%v [⭐️%v] %v", searchResult.FullName, searchResult.Stars, searchResult.Description)
		formattedSearchResults = append(formattedSearchResults, formatted)
	}

	return formattedSearchResults
}

// ValidateGitetatSearchQuery makes sure the Gitea search query is valid
func ValidateGiteaSearchQuery(searchQuery string) error {
	reSearch, err := regexp.Compile(constants.RegexGiteaSearch)
	if err != nil {
		return err
	}

	if !reSearch.MatchString(searchQuery) {
		return InvalidGiteaSearchQueryError{SearchQuery: searchQuery}
	}

	return nil
}
