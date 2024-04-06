package stew

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marwanhawari/stew/constants"
)

// LockFile contains all the data for the lockfile
type LockFile struct {
	Os       string        `json:"os"`
	Arch     string        `json:"arch"`
	Packages []PackageData `json:"packages"`
}

// PackageData contains the information for an installed binary
type PackageData struct {
	Source string `json:"source"`
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Tag    string `json:"tag"`
	Asset  string `json:"asset"`
	Binary string `json:"binary"`
	URL    string `json:"url"`
	Host   string `json:"host"`
}

func readLockFileJSON(lockFilePath string) (LockFile, error) {
	lockFileBytes, err := os.ReadFile(lockFilePath)
	if err != nil {
		return LockFile{}, err
	}

	var lockFile LockFile
	err = json.Unmarshal(lockFileBytes, &lockFile)
	if err != nil {
		return LockFile{}, err
	}

	return lockFile, nil
}

// WriteLockFileJSON will write the lockfile JSON file
func WriteLockFileJSON(lockFileJSON LockFile, outputPath string) error {
	lockFileBytes, err := json.MarshalIndent(lockFileJSON, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, lockFileBytes, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("📄 Updated %v\n", constants.GreenColor(outputPath))

	return nil
}

// RemovePackage will remove a package from a LockFile.Packages slice
func RemovePackage(pkgs []PackageData, index int) ([]PackageData, error) {
	if len(pkgs) == 0 {
		return []PackageData{}, NoPackagesInLockfileError{}
	}

	if index < 0 || index >= len(pkgs) {
		return []PackageData{}, IndexOutOfBoundsInLockfileError{}
	}

	return append(pkgs[:index], pkgs[index+1:]...), nil
}

// ReadStewfileContents will read the contents of the Stewfile
func ReadStewfileContents(stewfilePath string) ([]PackageData, error) {
	file, err := os.Open(stewfilePath)
	if err != nil {
		return []PackageData{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var packages []PackageData
	for scanner.Scan() {
		line := scanner.Text()
		packageAndOptions := strings.SplitN(line, "?", 2)
		packageString := packageAndOptions[0]
		options := make(map[string]string, 0)
		if len(packageAndOptions) == 2 {
			optionsSplit := strings.Split(packageAndOptions[1], "&")
			for _, option := range optionsSplit {
				optionkv := strings.Split(option, "=")
				options[optionkv[0]] = optionkv[1]
			}
		}
		if strings.HasPrefix(packageString, "https") || strings.HasPrefix(packageString, "http") {
			p := PackageData{
				URL:    packageString,
				Source: "other",
			}
			packages = append(packages, p)
			continue
		}
		splitInput := strings.Split(packageString, "@")
		owner := strings.Split(splitInput[0], "/")[0]
		repo := strings.Split(splitInput[0], "/")[1]
		p := PackageData{
			Repo:  repo,
			Owner: owner,
		}
		if len(splitInput) == 2 {
			tagAndAsset := strings.Split(splitInput[1], "#")
			if len(tagAndAsset) == 2 {
				p.Asset = tagAndAsset[1]
			}
			p.Tag = tagAndAsset[0]
		}
		p.Host = options["host"]
		p.Source = "github"
		if options["source"] != "" {
			p.Source = options["source"]
		}
		packages = append(packages, p)
	}

	if err := scanner.Err(); err != nil {
		return []PackageData{}, err
	}

	return packages, nil
}

func ReadStewLockFileContents(lockFilePath string) ([]PackageData, error) {
	lockFile, err := readLockFileJSON(lockFilePath)
	if err != nil {
		return []PackageData{}, err
	}

	return lockFile.Packages, nil
}

// NewLockFile creates a new instance of the LockFile struct
func NewLockFile(stewLockFilePath, userOS, userArch string) (LockFile, error) {
	var lockFile LockFile
	lockFileExists, err := PathExists(stewLockFilePath)
	if err != nil {
		return LockFile{}, err
	}
	if !lockFileExists {
		lockFile = LockFile{Os: userOS, Arch: userArch, Packages: []PackageData{}}
	} else {
		lockFile, err = readLockFileJSON(stewLockFilePath)
		if err != nil {
			return LockFile{}, err
		}
	}
	return lockFile, nil
}

// DeleteAssetAndBinary will delete the asset from the ~/.stew/pkg path and delete the binary from the ~/.stew/bin path
func DeleteAssetAndBinary(stewPkgPath, stewBinPath, asset, binary string) error {
	assetPath := filepath.Join(stewPkgPath, asset)
	binPath := filepath.Join(stewBinPath, binary)
	err := os.RemoveAll(assetPath)
	if err != nil {
		return err
	}
	err = os.RemoveAll(binPath)
	if err != nil {
		return err
	}
	return nil
}
