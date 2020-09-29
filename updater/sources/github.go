package sources

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/unly/wow-addon-updater/util"
)

// GithubSource is the source for addons and UIs hosted on github.com
type GithubSource struct {
	*source
	client    *github.Client
	repoRegex *regexp.Regexp
}

// NewGitHubSource returns a pointer to a newly created GithubSource.
func NewGitHubSource() *GithubSource {
	return &GithubSource{
		source:    newSource(regexp.MustCompile(`(https?://)?github\.com/([a-zA-Z0-9]|-)+/([a-zA-Z0-9]|-)+`), "github"),
		client:    github.NewClient(nil),
		repoRegex: regexp.MustCompile(`/([a-zA-Z0-9]|-)+/([a-zA-Z0-9]|-)+`),
	}
}

// GetLatestVersion returns the git tag of the latest release of the given repository URL.
func (g *GithubSource) GetLatestVersion(addonURL string) (string, error) {
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
func (g *GithubSource) DownloadAddon(addonURL, dir string) error {
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

	zipPath, err := g.downloadZip(url)
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

func (g *GithubSource) getLatestRelease(addonURL string) (*github.RepositoryRelease, error) {
	organization, repo, err := g.getOrgAndRepository(addonURL)
	if err != nil {
		return nil, err
	}

	release, resp, err := g.client.Repositories.GetLatestRelease(context.Background(), organization, repo)
	if err := checkHTTPResponse(resp.Response, err); err != nil {
		return nil, err
	}

	return release, nil
}

func (g *GithubSource) getOrgAndRepository(addonURL string) (string, string, error) {
	repo := g.repoRegex.FindString(addonURL)
	split := strings.Split(repo, "/")
	if len(split) != 3 || split[1] == "" || split[2] == "" {
		return "", "", fmt.Errorf("the given url %s is invalid for a github repository", addonURL)
	}

	return split[1], split[2], nil
}
