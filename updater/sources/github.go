package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v32/github"
)

var regex = regexp.MustCompile(`https://github.com/([a-zA-Z0-9]|-)+/([a-zA-Z0-9]|-)+`)
var repoRegex = regexp.MustCompile(`/([a-zA-Z0-9]|-)+/([a-zA-Z0-9]|-)+`)

type githubSource struct {
	client *github.Client
}

func NewGitHubSource() *githubSource {
	return &githubSource{
		client: github.NewClient(nil),
	}
}

func (g *githubSource) GetURLRegex() *regexp.Regexp {
	return regex
}

func (g *githubSource) GetLatestVersion(addon string) (string, error) {
	release, err := g.getLatestRelease(addon)
	if err != nil {
		return "", err
	}

	return release.GetTagName(), nil
}

func (g *githubSource) DownloadAddon(addon string) (string, error) {
	release, err := g.getLatestRelease(addon)
	if err != nil {
		return "", err
	}

	url := release.GetZipballURL()
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	org, repo, err := getOrgAndRepository(addon)
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("%s-%s.zip", org, repo)
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return path, err
}

func getOrgAndRepository(addon string) (string, string, error) {
	repo := repoRegex.FindString(addon)
	split := strings.Split(repo, "/")
	if len(split) != 3 || split[1] == "" || split[2] == "" {
		return "", "", fmt.Errorf("the given url is invalid for a github repository")
	}

	return split[1], split[2], nil
}

func (g *githubSource) getLatestRelease(addon string) (*github.RepositoryRelease, error) {
	organization, repo, err := getOrgAndRepository(addon)
	if err != nil {
		return nil, err
	}

	release, resp, err := g.client.Repositories.GetLatestRelease(context.Background(), organization, repo)
	if err != nil {
		return nil, err
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, fmt.Errorf("failed to get latest version. error code: %s", resp.Status)
	}

	return release, nil
}
