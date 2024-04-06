package stew

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/marwanhawari/stew/constants"
)

// GitlabProject contains information about the Gitlab project including Gitlab releases
type GitlabProject struct {
	Groups   []string
	Project  string
	Releases GitlabAPIResponse
}

// GitlabAPIResponse is the response from the Gitlab releases API
type GitlabAPIResponse []GitlabRelease

// GitlabRelease contains information about a Gitlab release, including the associated assets
type GitlabRelease struct {
	TagName string      `json:"tag_name"`
	Name    string      `json:"name"`
	Assets  GitlabAsset `json:"assets"`
}

// GitlabAsset contains information about a specific Gitlab asset
type GitlabAsset struct {
	Links []GitlabSources `json:"links"`
}

type GitlabSources struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DownloadURL string `json:"url"`
}

func readGitlabJSON(host string, groups []string, project, jsonString string) (GitlabAPIResponse, error) {
	var gtProject GitlabAPIResponse
	err := json.Unmarshal([]byte(jsonString), &gtProject)
	if err != nil {
		return GitlabAPIResponse{}, err
	}
	return gtProject, nil
}

func getGitlabJSON(host string, groups []string, project string) (string, error) {
	projectString := ""
	for _, group := range groups {
		projectString += group + "%2F"
	}
	projectString += project
	url := fmt.Sprintf("https://%s/api/v4/projects/%s/releases?per_page=100", host, projectString)
	fmt.Println(url)
	response, err := getHTTPResponseBody(url, "gitlab")
	if err != nil {
		return "", err
	}

	return response, nil
}

// NewGitlabProject creates a new instance of the GitlabProject struct
func NewGitlabProject(host string, groups []string, project string) (GitlabProject, error) {
	gtJSON, err := getGitlabJSON(host, groups, project)
	if err != nil {
		return GitlabProject{}, err
	}

	gtAPIResponse, err := readGitlabJSON(host, groups, project, gtJSON)
	if err != nil {
		return GitlabProject{}, err
	}

	ghProject := GitlabProject{Groups: groups, Project: project, Releases: gtAPIResponse}

	return ghProject, nil
}

// GetGitlabReleasesTags gets a string slice of the releases for a GitlabProject
func GetGitlabReleasesTags(ghProject GitlabProject, host string) ([]string, error) {
	releasesTags := []string{}

	for _, release := range ghProject.Releases {
		releasesTags = append(releasesTags, release.TagName)
	}

	err := gitlabReleasesFound(releasesTags, ghProject.Groups, ghProject.Project, host)
	if err != nil {
		return []string{}, err
	}

	return releasesTags, nil
}

func gitlabReleasesFound(releaseTags []string, owner []string, repo string, host string) error {
	if len(releaseTags) == 0 {
		return GitlabReleasesNotFoundError{Groups: owner, Repo: repo, Host: host}
	}
	return nil
}

// GetGitlabReleasesAssets gets a string slice of the assets for a GitlabRelease
func GetGitlabReleasesAssets(ghProject GitlabProject, tag string) ([]string, error) {
	releaseAssets := []string{}

	for _, release := range ghProject.Releases {
		if release.TagName == tag {
			for _, asset := range release.Assets.Links {
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
