package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marwanhawari/stew/constants"
	stew "github.com/marwanhawari/stew/lib"
)

// Install is executed when you run `stew install`
func Install(host, hostType string, cliInputs []string) {
	var err error

	userOS, userArch, _, systemInfo, err := stew.Initialize()
	stew.CatchAndExit(err)

	for _, cliInput := range cliInputs {
		if strings.Contains(cliInput, "Stewfile.lock.json") {
			packages, err := stew.ReadStewLockFileContents(cliInput)
			stew.CatchAndExit(err)
			for _, packageData := range packages {
				switch packageData.Source {
				case "other":
					Install("", "", []string{packageData.URL})
				case "gitea":
					Install(
						packageData.Host,
						"gitea",
						[]string{
							packageData.Owner + "/" + packageData.Repo + "@" + packageData.Tag + "#" + packageData.Asset,
						},
					)
				default:
					Install(
						"",
						"github",
						[]string{
							packageData.Owner + "/" + packageData.Repo + "@" + packageData.Tag + "#" + packageData.Asset,
						},
					)
				}
			}
			return
		}

		if strings.Contains(cliInput, "Stewfile") {
			packages, err := stew.ReadStewfileContents(cliInput)
			stew.CatchAndExit(err)
			for _, packageData := range packages {
				fmt.Printf("%+v\n", packageData)
				switch packageData.Source {
				case "other":
					Install("", "", []string{packageData.URL})
				case "gitea":
					Install(
						packageData.Host,
						"gitea",
						[]string{
							packageData.Owner + "/" + packageData.Repo + "@" + packageData.Tag + "#" + packageData.Asset,
						},
					)
				default:
					Install(
						"",
						"github",
						[]string{
							packageData.Owner + "/" + packageData.Repo + "@" + packageData.Tag + "#" + packageData.Asset,
						},
					)
				}
			}
			return
		}
	}

	if len(cliInputs) == 0 {
		return
	}

	for _, cliInput := range cliInputs {
		sp := constants.LoadingSpinner

		stewBinPath := systemInfo.StewBinPath
		stewPkgPath := systemInfo.StewPkgPath
		stewLockFilePath := systemInfo.StewLockFilePath
		stewTmpPath := systemInfo.StewTmpPath

		parsedInput, err := stew.ParseCLIInput(cliInput)
		stew.CatchAndExit(err)

		owner := parsedInput.Owner
		repo := parsedInput.Repo
		tag := parsedInput.Tag
		asset := parsedInput.Asset
		downloadURL := parsedInput.DownloadURL

		lockFile, err := stew.NewLockFile(stewLockFilePath, userOS, userArch)
		stew.CatchAndExit(err)

		err = os.RemoveAll(stewTmpPath)
		stew.CatchAndExit(err)
		err = os.MkdirAll(stewTmpPath, 0755)
		stew.CatchAndExit(err)

		if parsedInput.IsGithubInput {
			switch hostType {
			case "gitea":
				fmt.Println(constants.GreenColor(owner + "/" + repo))
				sp.Start()
				giteaProject, err := stew.NewGiteaProject(host, owner, repo)
				sp.Stop()
				stew.CatchAndExit(err)

				// This will make sure that there are any tags at all
				releaseTags, err := stew.GetGiteaReleasesTags(giteaProject)
				stew.CatchAndExit(err)

				if tag == "" || tag == "latest" {
					tag = giteaProject.Releases[0].TagName
				}

				// Need to make sure user input tag is in the tags
				tagIndex, tagFound := stew.Contains(releaseTags, tag)
				if !tagFound {
					tag, err = stew.WarningPromptSelect(
						fmt.Sprintf(
							"Could not find a release with the tag %v - please select a release:",
							constants.YellowColor(tag),
						),
						releaseTags,
					)
					stew.CatchAndExit(err)
					tagIndex, _ = stew.Contains(releaseTags, tag)
				}

				// Make sure there are any assets at all
				releaseAssets, err := stew.GetGiteaReleasesAssets(giteaProject, tag)
				stew.CatchAndExit(err)

				if asset == "" {
					asset, err = stew.DetectAsset(userOS, userArch, releaseAssets)
				}
				stew.CatchAndExit(err)

				assetIndex, assetFound := stew.Contains(releaseAssets, asset)
				if !assetFound {
					asset, err = stew.WarningPromptSelect(
						fmt.Sprintf("Could not find the asset %v - please select an asset:", constants.YellowColor(asset)),
						releaseAssets,
					)
					stew.CatchAndExit(err)
					assetIndex, _ = stew.Contains(releaseAssets, asset)
				}
				downloadURL = giteaProject.Releases[tagIndex].Assets[assetIndex].DownloadURL
				owner = giteaProject.Owner
				repo = giteaProject.Repo
			default:
				hostType = "github"
				fmt.Println(constants.GreenColor(owner + "/" + repo))
				sp.Start()
				githubProject, err := stew.NewGithubProject(owner, repo)
				sp.Stop()
				stew.CatchAndExit(err)

				// This will make sure that there are any tags at all
				releaseTags, err := stew.GetGithubReleasesTags(githubProject)
				stew.CatchAndExit(err)

				if tag == "" || tag == "latest" {
					tag = githubProject.Releases[0].TagName
				}

				// Need to make sure user input tag is in the tags
				tagIndex, tagFound := stew.Contains(releaseTags, tag)
				if !tagFound {
					tag, err = stew.WarningPromptSelect(
						fmt.Sprintf(
							"Could not find a release with the tag %v - please select a release:",
							constants.YellowColor(tag),
						),
						releaseTags,
					)
					stew.CatchAndExit(err)
					tagIndex, _ = stew.Contains(releaseTags, tag)
				}

				// Make sure there are any assets at all
				releaseAssets, err := stew.GetGithubReleasesAssets(githubProject, tag)
				stew.CatchAndExit(err)

				if asset == "" {
					asset, err = stew.DetectAsset(userOS, userArch, releaseAssets)
				}
				stew.CatchAndExit(err)

				assetIndex, assetFound := stew.Contains(releaseAssets, asset)
				if !assetFound {
					asset, err = stew.WarningPromptSelect(
						fmt.Sprintf("Could not find the asset %v - please select an asset:", constants.YellowColor(asset)),
						releaseAssets,
					)
					stew.CatchAndExit(err)
					assetIndex, _ = stew.Contains(releaseAssets, asset)
				}
				downloadURL = githubProject.Releases[tagIndex].Assets[assetIndex].DownloadURL
				owner = githubProject.Owner
				repo = githubProject.Repo
			}
		} else {
			fmt.Println(constants.GreenColor(asset))
		}
		downloadPath := filepath.Join(stewPkgPath, asset)
		err = stew.DownloadFile(downloadPath, downloadURL, hostType)
		stew.CatchAndExit(err)
		fmt.Printf("✅ Downloaded %v to %v\n", constants.GreenColor(asset), constants.GreenColor(stewPkgPath))

		binaryName, err := stew.InstallBinary(downloadPath, repo, systemInfo, &lockFile, false)
		if err != nil {
			os.RemoveAll(downloadPath)
			stew.CatchAndExit(err)
		}

		var packageData stew.PackageData
		if parsedInput.IsGithubInput {
			packageData = stew.PackageData{
				Source: hostType,
				Owner:  owner,
				Repo:   repo,
				Tag:    tag,
				Asset:  asset,
				Binary: binaryName,
				URL:    downloadURL,
				Host:   host,
			}
		} else {
			packageData = stew.PackageData{
				Source: "other",
				Owner:  "",
				Repo:   "",
				Tag:    "",
				Asset:  asset,
				Binary: binaryName,
				URL:    downloadURL,
			}
		}

		lockFile.Packages = append(lockFile.Packages, packageData)

		err = stew.WriteLockFileJSON(lockFile, stewLockFilePath)
		stew.CatchAndExit(err)

		fmt.Printf(
			"✨ Successfully installed the %v binary in %v\n",
			constants.GreenColor(binaryName),
			constants.GreenColor(stewBinPath),
		)

	}
}
