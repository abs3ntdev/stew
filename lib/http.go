package stew

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func getHTTPResponseBody(url string, hostType string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	switch hostType {
	case "github":
		req.Header.Add("Accept", "application/octet-stream")
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken != "" {
			req.Header.Add("Authorization", fmt.Sprintf("token %v", githubToken))
		}
	case "gitea":
		req.Header.Add("Accept", "application/octet-stream")
		giteaToken := os.Getenv("GITEA_TOKEN")
		if giteaToken != "" {
			req.Header.Add("Authorization", fmt.Sprintf("token %v", giteaToken))
		}
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", NonZeroStatusCodeError{res.StatusCode}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
