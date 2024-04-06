package stew

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func getHTTPResponseBody(urlInput string, hostType string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlInput, nil)
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
	case "gitlab":
		req.Header.Add("Accept", "application/octet-stream")
		parsedUrl, err := url.Parse(urlInput)
		CatchAndExit(err)
		host := parsedUrl.Host
		host = strings.ReplaceAll(host, ".", "_")
		host = strings.ToUpper(host)
		giteaToken := os.Getenv(host + "_TOKEN")
		if giteaToken != "" {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", giteaToken))
		}
	case "gitea":
		req.Header.Add("Accept", "application/octet-stream")
		parsedUrl, err := url.Parse(urlInput)
		CatchAndExit(err)
		host := parsedUrl.Host
		host = strings.ReplaceAll(host, ".", "_")
		host = strings.ToUpper(host)
		giteaToken := os.Getenv(host + "_TOKEN")
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
