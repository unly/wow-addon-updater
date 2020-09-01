package sources

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/unly/wow-addon-updater/util"
)

var regex = regexp.MustCompile(`https://github.com/([a-zA-Z0-9]|-)+/([a-zA-Z0-9]|-)+`)
var repoRegex = regexp.MustCompile(`/([a-zA-Z0-9]|-)+/([a-zA-Z0-9]|-)+`)
var tempDir string = createTempDir()

type GithubSource struct {
	client *github.Client
}

func NewGitHubSource() *GithubSource {
	return &GithubSource{
		client: github.NewClient(nil),
	}
}

func (g *GithubSource) GetURLRegex() *regexp.Regexp {
	return regex
}

func (g *GithubSource) GetLatestVersion(addon string) (string, error) {
	release, err := g.getLatestRelease(addon)
	if err != nil {
		return "", err
	}

	return release.GetTagName(), nil
}

func (g *GithubSource) DownloadAddon(addon, dir string) error {
	release, err := g.getLatestRelease(addon)
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

	zipPath, err := downloadZip(url, tempDir)
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
		_, repo, err := getOrgAndRepository(addon)
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

func (g *GithubSource) getLatestRelease(addon string) (*github.RepositoryRelease, error) {
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

func downloadZip(url, dir string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile(dir, "*.zip")
	if err != nil {
		return "", err
	}

	path := file.Name()
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

func createTempDir() string {
	path, err := ioutil.TempDir("", "wow-updater-github")
	if err != nil {
		panic(err)
	}

	return path
}
