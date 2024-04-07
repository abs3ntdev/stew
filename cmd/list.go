package cmd

import (
	"encoding/json"
	"fmt"

	stew "github.com/marwanhawari/stew/lib"
)

// List is executed when you run `stew list`
func List(cliTagsFlag bool) {
	userOS, userArch, _, systemInfo, err := stew.Initialize()
	stew.CatchAndExit(err)

	stewLockFilePath := systemInfo.StewLockFilePath

	lockFile, err := stew.NewLockFile(stewLockFilePath, userOS, userArch)
	stew.CatchAndExit(err)

	if len(lockFile.Packages) == 0 {
		return
	}

	sources := make(map[string][]string)
	for _, pkg := range lockFile.Packages {
		switch pkg.Source {
		case "other":
			sources["urls"] = append(sources["other"], pkg.URL)
			fmt.Println(pkg.URL)
		case "github":
			if cliTagsFlag {
				sources["github.com"] = append(sources["github.com"], pkg.Owner+"/"+pkg.Repo+"@"+pkg.Tag)
			} else {
				sources["github.com"] = append(sources["github.com"], pkg.Owner+"/"+pkg.Repo)
			}
		case "gitlab", "gitea":
			if cliTagsFlag {
				sources[pkg.Host] = append(sources[pkg.Host], pkg.Owner+"/"+pkg.Repo+"@"+pkg.Tag)
			} else {
				sources[pkg.Host] = append(sources[pkg.Host], pkg.Owner+"/"+pkg.Repo)
			}
		}
	}
	out, err := json.MarshalIndent(sources, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(out))
}
