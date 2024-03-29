package github

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v38/github"

	"github.com/unly/wow-addon-updater/updater"
	"github.com/unly/wow-addon-updater/updater/sources"
	"github.com/unly/wow-addon-updater/util"
)

var (
	regex     = regexp.MustCompile(`^(https?://)?github\.com/([a-zA-Z0-9]|-)+/([a-zA-Z0-9]|-)+/?$`)
	repoRegex = regexp.MustCompile(`/([a-zA-Z0-9]|-)+/([a-zA-Z0-9]|-)+`)
)

//go:generate go run github.com/vektra/mockery/v2 --case=underscore  --name=githubAPI --structname=MockGitHubAPI

type githubAPI interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
}

type githubSource struct {
	downloader sources.Downloader
	client     *http.Client
	api        githubAPI
}

// New returns a pointer to a newly created GithubSource.
func New(client *http.Client) (updater.UpdateSource, error) {
	if client == nil {
		client = http.DefaultClient
	}
	d, err := sources.NewDownloader(client)
	if err != nil {
		return nil, err
	}

	return &githubSource{
		downloader: d,
		client:     client,
		api:        github.NewClient(client).Repositories,
	}, nil
}

func (githubSource) GetURLRegex() *regexp.Regexp {
	return regex
}

// GetLatestVersion returns the git tag of the latest release of the given repository URL.
func (g *githubSource) GetLatestVersion(addonURL string) (string, error) {
	release, err := g.getLatestRelease(addonURL)
	if err != nil {
		return "", err
	}

	return release.GetTagName(), nil
}

// DownloadAddon downloads and unzip the addon if there is just one .zip archive attached
// to the latest release.
// Otherwise the git repository itself will be downloaded and copied to the given
// directory.
func (g *githubSource) DownloadAddon(addonURL, dir string) error {
	release, err := g.getLatestRelease(addonURL)
	if err != nil {
		return err
	}

	url := release.GetZipballURL()
	gitArchive := true
	// approach to download an asset rather than the entire repository
	if len(release.Assets) == 1 {
		asset := release.Assets[0]
		if asset.GetContentType() == "application/x-zip-compressed" {
			url = asset.GetBrowserDownloadURL()
			gitArchive = false
		}
	}

	zipPath, err := g.downloader.DownloadZip(url)
	if err != nil {
		return err
	}

	files, err := util.Unzip(zipPath, dir)
	if err != nil {
		return err
	}

	// rename the git archive to the repository name
	if gitArchive {
		minLength := math.MaxInt32
		dirs := make([]string, 0)
		for _, f := range files {
			length := len(f)
			if length < minLength {
				dirs = make([]string, 0)
				minLength = length
			}
			if length == minLength {
				dirs = append(dirs, f)
			}
		}
		if len(dirs) != 1 {
			return fmt.Errorf("the git archive does not have a single root directory")
		}
		_, repo, err := g.getOrgAndRepository(addonURL)
		if err != nil {
			return err
		}
		//TODO overwrite existing files
		err = os.Rename(dirs[0], filepath.Join(dir, repo))
		if err != nil {
			return err
		}
	}

	return err
}

func (g *githubSource) getLatestRelease(addonURL string) (*github.RepositoryRelease, error) {
	organization, repo, err := g.getOrgAndRepository(addonURL)
	if err != nil {
		return nil, err
	}

	release, resp, err := g.api.GetLatestRelease(context.Background(), organization, repo)
	if err != nil {
		return nil, err
	}
	if err := util.CheckHTTPResponse(resp.Response, err); err != nil {
		return nil, err
	}

	return release, nil
}

func (g *githubSource) getOrgAndRepository(addonURL string) (string, string, error) {
	repo := repoRegex.FindString(addonURL)
	split := strings.Split(repo, "/")
	if len(split) != 3 || split[1] == "" || split[2] == "" {
		return "", "", fmt.Errorf("the given url %s is invalid for a github repository", addonURL)
	}

	return split[1], split[2], nil
}

func (g *githubSource) Close() error {
	return g.downloader.Close()
}
