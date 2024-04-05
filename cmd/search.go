package cmd

import (
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/marwanhawari/stew/constants"
	stew "github.com/marwanhawari/stew/lib"
)

// Search is executed when you run `stew search`
func Search(host, hostType, searchQuery string) {
	if hostType == "" {
		hostType = "github"
	}
	sp := constants.LoadingSpinner

	err := stew.ValidateCLIInput(searchQuery)
	stew.CatchAndExit(err)

	err = stew.ValidateGithubSearchQuery(searchQuery)
	stew.CatchAndExit(err)

	var searchResults stew.RepoSearch
	sp.Start()
	switch hostType {
	case "gitea":
		if host == "" {
			stew.CatchAndExit(fmt.Errorf("Host is required for Gitea search"))
		}
		searchResults, err = stew.NewGiteaSearch(host, searchQuery)
	default:
		searchResults, err = stew.NewGithubSearch(searchQuery)
	}
	sp.Stop()
	stew.CatchAndExit(err)

	if len(searchResults.Items) == 0 {
		stew.CatchAndExit(stew.NoGithubSearchResultsError{SearchQuery: searchResults.SearchQuery})
	}

	formattedSearchResults := stew.FormatSearchResults(searchResults)

	githubProjectName, err := stew.PromptSelect(
		fmt.Sprintf("Choose a %s project:", cases.Title(language.English, cases.Compact).String(hostType)),
		formattedSearchResults,
	)
	stew.CatchAndExit(err)

	searchResultIndex, _ := stew.Contains(formattedSearchResults, githubProjectName)

	Browse(host, hostType, searchResults.Items[searchResultIndex].FullName)
}
