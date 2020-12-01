package updater

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/config"
	"github.com/unly/wow-addon-updater/updater/mocks"
	"github.com/unly/wow-addon-updater/util"
	"github.com/unly/wow-addon-updater/util/tests/helpers"

	"gopkg.in/yaml.v3"
)

func Test_readVersionsFile(t *testing.T) {
	type readVersionsFileTest struct {
		path          string
		errorExpected bool
		want          versions
		teardown      helpers.TearDown
	}

	tests := []func() *readVersionsFileTest{
		func() *readVersionsFileTest {
			v := versions{
				Classic: []addon{
					{
						Name:    "Hello",
						Version: "World",
					},
				},
				Retail: []addon{},
			}
			content, err := yaml.Marshal(v)
			assert.NoError(t, err)
			file := helpers.TempFile(t, "", content)

			return &readVersionsFileTest{
				path:          file,
				errorExpected: false,
				want: versions{
					Classic: []addon{
						{
							Name:    "Hello",
							Version: "World",
						},
					},
					Retail: []addon{},
				},
				teardown: helpers.DeleteFile(t, file),
			}
		},
		func() *readVersionsFileTest {
			v := versions{}
			content, err := yaml.Marshal(v)
			assert.NoError(t, err)
			file := helpers.TempFile(t, "", content)

			return &readVersionsFileTest{
				path:          file,
				errorExpected: false,
				want: versions{
					Classic: []addon{},
					Retail:  []addon{},
				},
				teardown: helpers.DeleteFile(t, file),
			}
		},
		func() *readVersionsFileTest {
			file := helpers.TempFile(t, "", []byte("blub"))

			return &readVersionsFileTest{
				path:          file,
				errorExpected: true,
				want:          versions{},
				teardown:      helpers.DeleteFile(t, file),
			}
		},
		func() *readVersionsFileTest {
			return &readVersionsFileTest{
				path:          "not existing",
				errorExpected: false,
				want:          versions{},
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *readVersionsFileTest {
			return &readVersionsFileTest{
				path:          ".",
				errorExpected: false,
				want:          versions{},
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *readVersionsFileTest {
			return &readVersionsFileTest{
				path:          "",
				errorExpected: false,
				want:          versions{},
				teardown:      helpers.NoopTeardown(),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()
		actual, err := readVersionsFile(tt.path)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		}

		tt.teardown()
	}
}

func Test_saveVersionsFile(t *testing.T) {
	type saveVersionsFileTest struct {
		updater       *Updater
		errorExpected bool
		want          versions
		teardown      helpers.TearDown
	}

	tests := []func() *saveVersionsFileTest{
		func() *saveVersionsFileTest {
			file, err := util.HideFile(helpers.TempFile(t, "", []byte{}))
			assert.NoError(t, err)

			return &saveVersionsFileTest{
				updater: &Updater{
					versionFile: file,
				},
				errorExpected: false,
				want: versions{
					Classic: []addon{},
					Retail:  []addon{},
				},
				teardown: helpers.DeleteFile(t, file),
			}
		},
		func() *saveVersionsFileTest {
			file, err := util.HideFile(helpers.TempFile(t, "", []byte{}))
			assert.NoError(t, err)

			return &saveVersionsFileTest{
				updater: &Updater{
					classic: gameUpdater{
						versions: map[string]addon{
							"addon1": {
								Name:    "addon1",
								Version: "1",
							},
							"addon2": {
								Name:    "addon2",
								Version: "2",
							},
						},
					},
					retail: gameUpdater{
						versions: map[string]addon{
							"addon3": {
								Name:    "addon3",
								Version: "3",
							},
						},
					},
					versionFile: file,
				},
				errorExpected: false,
				want: versions{
					Classic: []addon{
						{
							Name:    "addon1",
							Version: "1",
						},
						{
							Name:    "addon2",
							Version: "2",
						},
					},
					Retail: []addon{
						{
							Name:    "addon3",
							Version: "3",
						},
					},
				},
				teardown: helpers.DeleteFile(t, file),
			}
		},
		func() *saveVersionsFileTest {
			return &saveVersionsFileTest{
				updater:       &Updater{},
				errorExpected: true,
				want:          versions{},
				teardown:      helpers.NoopTeardown(),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		err := saveVersionsFile(tt.updater)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			out, err := ioutil.ReadFile(tt.updater.versionFile)
			assert.NoError(t, err)
			var actual versions
			err = yaml.Unmarshal(out, &actual)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.want.Classic, actual.Classic)
			assert.ElementsMatch(t, tt.want.Retail, actual.Retail)
		}

		tt.teardown()
	}
}

func Test_mapAddonVersions(t *testing.T) {
	tests := []struct {
		addons []addon
		want   map[string]addon
	}{
		{
			addons: []addon{
				{
					Name:    "addon1",
					Version: "1.2.3",
				},
				{
					Name: "addon2",
				},
			},
			want: map[string]addon{
				"addon1": {
					Name:    "addon1",
					Version: "1.2.3",
				},
				"addon2": {
					Name: "addon2",
				},
			},
		},
		{
			addons: []addon{
				{
					Name:    "addon1",
					Version: "1.2.3",
				},
				{},
			},
			want: map[string]addon{
				"addon1": {
					Name:    "addon1",
					Version: "1.2.3",
				},
				"": {},
			},
		},
		{
			addons: []addon{},
			want:   map[string]addon{},
		},
		{
			addons: nil,
			want:   map[string]addon{},
		},
	}

	for _, tt := range tests {
		actual := mapAddonVersions(tt.addons)

		assert.Equal(t, tt.want, actual)
	}
}

func Test_getAddons(t *testing.T) {
	tests := []struct {
		updater *gameUpdater
		want    []addon
	}{
		{
			updater: &gameUpdater{},
			want:    []addon{},
		},
		{
			updater: &gameUpdater{
				versions: map[string]addon{
					"addon1": {
						Name: "addon1",
					},
					"addon2": {
						Name:    "addon2",
						Version: "1.2.3",
					},
				},
			},
			want: []addon{
				{
					Name: "addon1",
				},
				{
					Name:    "addon2",
					Version: "1.2.3",
				},
			},
		},
	}

	for _, tt := range tests {
		actual := getAddons(tt.updater)

		assert.ElementsMatch(t, tt.want, actual)
	}
}

func Test_getCurrentVersion(t *testing.T) {
	tests := []struct {
		updater *gameUpdater
		addon   string
		want    string
	}{
		{
			updater: &gameUpdater{
				versions: map[string]addon{
					"addon": {
						Name:    "addon",
						Version: "1.2.3",
					},
				},
			},
			addon: "addon",
			want:  "1.2.3",
		},
		{
			updater: &gameUpdater{
				versions: map[string]addon{
					"addon": {
						Name:    "addon",
						Version: "1.2.3",
					},
				},
			},
			addon: "not existing",
			want:  "",
		},
		{
			updater: &gameUpdater{
				versions: map[string]addon{
					"addon": {
						Name: "addon",
					},
				},
			},
			addon: "addon",
			want:  "",
		},
		{
			updater: &gameUpdater{},
			addon:   "addon",
			want:    "",
		},
	}

	for _, tt := range tests {
		actual := tt.updater.getCurrentVersion(tt.addon)

		assert.Equal(t, tt.want, actual)
	}
}

func Test_setCurrentVersion(t *testing.T) {
	tests := []struct {
		updater *gameUpdater
		addon   string
		version string
	}{
		{
			updater: &gameUpdater{
				versions: map[string]addon{
					"addon": {
						Name:    "addon",
						Version: "1.2.3",
					},
				},
			},
			addon:   "addon",
			version: "1.2.4",
		},
		{
			updater: &gameUpdater{
				versions: map[string]addon{
					"addon": {
						Name:    "addon",
						Version: "1.2.3",
					},
				},
			},
			addon:   "addon",
			version: "",
		},
		{
			updater: &gameUpdater{
				versions: map[string]addon{
					"addon": {
						Name:    "addon",
						Version: "1.2.3",
					},
				},
			},
			addon:   "not existing",
			version: "1.2.4",
		},
		{
			updater: &gameUpdater{},
			addon:   "new",
			version: "1",
		},
	}

	for _, tt := range tests {
		tt.updater.setCurrentVersion(tt.addon, tt.version)

		actual, ok := tt.updater.versions[tt.addon]
		assert.True(t, ok)
		assert.Equal(t, tt.version, actual.Version)
	}
}

func Test_NewUpdater(t *testing.T) {
	type newUpdaterTest struct {
		config        config.Config
		sources       []UpdateSource
		versionFile   string
		errorExpected bool
		want          *Updater
		teardown      helpers.TearDown
	}

	tests := []func() *newUpdaterTest{
		func() *newUpdaterTest {
			return &newUpdaterTest{
				config:        config.Config{},
				sources:       nil,
				versionFile:   ".file",
				errorExpected: false,
				want: &Updater{
					classic: gameUpdater{
						versions: map[string]addon{},
					},
					retail: gameUpdater{
						versions: map[string]addon{},
					},
					sources:     []UpdateSource{},
					versionFile: ".file",
				},
				teardown: helpers.NoopTeardown(),
			}
		},
		func() *newUpdaterTest {
			c := config.Config{
				Classic: config.WowConfig{
					Path: "path/to/addons/dir",
					AddOns: []string{
						"addon1",
						"addon2",
					},
				},
			}
			sources := []UpdateSource{
				&mocks.MockUpdateSource{},
			}

			return &newUpdaterTest{
				config:        c,
				sources:       sources,
				versionFile:   ".file",
				errorExpected: false,
				want: &Updater{
					classic: gameUpdater{
						config:   c.Classic,
						versions: map[string]addon{},
					},
					retail: gameUpdater{
						versions: map[string]addon{},
					},
					sources:     sources,
					versionFile: ".file",
				},
				teardown: helpers.NoopTeardown(),
			}
		},
		func() *newUpdaterTest {
			v := versions{
				Classic: []addon{
					{
						Name:    "addon1",
						Version: "1.2.3",
					},
				},
				Retail: []addon{
					{
						Name: "addon2",
					},
				},
			}
			content, err := json.Marshal(&v)
			assert.NoError(t, err)
			file, err := util.HideFile(helpers.TempFile(t, "", content))
			assert.NoError(t, err)
			c := config.Config{
				Classic: config.WowConfig{
					Path: "path/to/addons/dir",
					AddOns: []string{
						"addon1",
						"addon2",
					},
				},
				Retail: config.WowConfig{
					Path: "path/to/retail/addons/dir",
					AddOns: []string{
						"addon3",
						"addon4",
					},
				},
			}

			return &newUpdaterTest{
				config:        c,
				sources:       []UpdateSource{},
				versionFile:   file,
				errorExpected: false,
				want: &Updater{
					classic: gameUpdater{
						config:   c.Classic,
						versions: map[string]addon{},
					},
					retail: gameUpdater{
						config:   c.Retail,
						versions: map[string]addon{},
					},
					sources:     []UpdateSource{},
					versionFile: file,
				},
				teardown: helpers.DeleteFile(t, file),
			}
		},
		func() *newUpdaterTest {
			file := helpers.TempFile(t, "", []byte("just text"))

			return &newUpdaterTest{
				config:        config.Config{},
				sources:       []UpdateSource{},
				versionFile:   file,
				errorExpected: true,
				want:          nil,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()
		actual, err := NewUpdater(tt.config, tt.sources, tt.versionFile)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		}

		tt.teardown()
	}
}

func Test_UpdateAddons(t *testing.T) {
	type updateAddonsTest struct {
		updater           *Updater
		errorExpected     bool
		versionFileExists bool
		teardown          helpers.TearDown
	}

	tests := []func() *updateAddonsTest{
		func() *updateAddonsTest {
			file, err := util.HideFile(helpers.TempFile(t, "", []byte{}))
			assert.NoError(t, err)

			return &updateAddonsTest{
				updater: &Updater{
					versionFile: file,
				},
				errorExpected:     false,
				versionFileExists: true,
				teardown:          helpers.DeleteFile(t, file),
			}
		},
		func() *updateAddonsTest {
			return &updateAddonsTest{
				updater:           &Updater{},
				errorExpected:     false,
				versionFileExists: false,
				teardown:          helpers.NoopTeardown(),
			}
		},
		func() *updateAddonsTest {
			file := helpers.TempFile(t, "", []byte{})

			return &updateAddonsTest{
				updater: &Updater{
					classic: gameUpdater{
						config: config.WowConfig{
							AddOns: []string{
								"addon",
							},
						},
					},
					sources:     []UpdateSource{},
					versionFile: file,
				},
				errorExpected:     true,
				versionFileExists: true,
				teardown:          helpers.DeleteFile(t, file),
			}
		},
		func() *updateAddonsTest {
			file := helpers.TempFile(t, "", []byte{})

			return &updateAddonsTest{
				updater: &Updater{
					retail: gameUpdater{
						config: config.WowConfig{
							AddOns: []string{
								"addon",
							},
						},
					},
					sources:     []UpdateSource{},
					versionFile: file,
				},
				errorExpected:     true,
				versionFileExists: true,
				teardown:          helpers.DeleteFile(t, file),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		err := tt.updater.UpdateAddons()

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			if tt.versionFileExists {
				assert.FileExists(t, tt.updater.versionFile)
			} else {
				assert.NoFileExists(t, tt.updater.versionFile)
			}
		}

		tt.teardown()
	}
}

func Test_getSource(t *testing.T) {
	type getSourceTest struct {
		sources       []UpdateSource
		addonURL      string
		errorExpected bool
		want          UpdateSource
	}

	tests := []func() *getSourceTest{
		func() *getSourceTest {
			return &getSourceTest{
				sources:       nil,
				addonURL:      "example.com",
				errorExpected: true,
				want:          nil,
			}
		},
		func() *getSourceTest {
			m := mocks.MockUpdateSource{}
			m.On("GetURLRegex").Return(regexp.MustCompile("test.com/.+"))

			return &getSourceTest{
				sources: []UpdateSource{
					&m,
				},
				addonURL:      "example.com",
				errorExpected: true,
				want:          nil,
			}
		},
		func() *getSourceTest {
			m1 := mocks.MockUpdateSource{}
			m1.On("GetURLRegex").Return(regexp.MustCompile("test.com/.+"))
			m2 := mocks.MockUpdateSource{}
			m2.On("GetURLRegex").Return(regexp.MustCompile("example.com/.+"))

			return &getSourceTest{
				sources: []UpdateSource{
					&m1,
					&m2,
				},
				addonURL:      "example.com/addon",
				errorExpected: false,
				want:          &m2,
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		actual, err := getSource(tt.sources, tt.addonURL)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		}
	}
}

func Test_updateAddon(t *testing.T) {
	type updateAddonTest struct {
		updater       *gameUpdater
		addonURL      string
		source        UpdateSource
		errorExpected bool
	}

	tests := []func() *updateAddonTest{
		func() *updateAddonTest {
			url := "example.com/addon"
			g := &gameUpdater{
				config: config.WowConfig{
					Path: "addon/dir",
				},
			}
			m := mocks.MockUpdateSource{}
			m.On("GetLatestVersion", url).Return("1.2.3", nil)
			m.On("DownloadAddon", url, g.config.Path).Return(nil)

			return &updateAddonTest{
				updater:       g,
				addonURL:      url,
				source:        &m,
				errorExpected: false,
			}
		},
		func() *updateAddonTest {
			url := "example.com/addon"
			m := mocks.MockUpdateSource{}
			m.On("GetLatestVersion", url).Return("", errors.New("failed to get latest version"))

			return &updateAddonTest{
				updater:       &gameUpdater{},
				addonURL:      url,
				source:        &m,
				errorExpected: true,
			}
		},
		func() *updateAddonTest {
			url := "example.com/addon"
			g := &gameUpdater{
				config: config.WowConfig{
					Path: "addon/dir",
				},
			}
			m := mocks.MockUpdateSource{}
			m.On("GetLatestVersion", url).Return("1.2.3", nil)
			m.On("DownloadAddon", url, g.config.Path).Return(errors.New("failed to download addon"))

			return &updateAddonTest{
				updater:       g,
				addonURL:      url,
				source:        &m,
				errorExpected: true,
			}
		},
		func() *updateAddonTest {
			url := "example.com/addon"
			g := &gameUpdater{
				config: config.WowConfig{
					Path: "addon/dir",
				},
				versions: map[string]addon{
					url: {
						Version: "1.2.1",
					},
				},
			}
			m := mocks.MockUpdateSource{}
			m.On("GetLatestVersion", url).Return("1.2.3", nil)
			m.On("DownloadAddon", url, g.config.Path).Return(nil)

			return &updateAddonTest{
				updater:       g,
				addonURL:      url,
				source:        &m,
				errorExpected: false,
			}
		},
		func() *updateAddonTest {
			url := "example.com/addon"
			g := &gameUpdater{
				config: config.WowConfig{
					Path: "addon/dir",
				},
				versions: map[string]addon{
					url: {
						Version: "1.2.3",
					},
				},
			}
			m := mocks.MockUpdateSource{}
			m.On("GetLatestVersion", url).Return("1.2.3", nil)

			return &updateAddonTest{
				updater:       g,
				addonURL:      url,
				source:        &m,
				errorExpected: false,
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		err := tt.updater.updateAddon(tt.addonURL, tt.source)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)

			addon, ok := tt.updater.versions[tt.addonURL]
			assert.True(t, ok)
			want, err := tt.source.GetLatestVersion(tt.addonURL)
			assert.NoError(t, err)
			assert.Equal(t, want, addon.Version)
		}
	}
}

func Test_updateAddons(t *testing.T) {
	type updateAddons struct {
		updater       *gameUpdater
		sources       []UpdateSource
		errorExpected bool
	}

	tests := []func() *updateAddons{
		func() *updateAddons {
			return &updateAddons{
				updater:       &gameUpdater{},
				sources:       nil,
				errorExpected: false,
			}
		},
		func() *updateAddons {
			return &updateAddons{
				updater: &gameUpdater{
					config: config.WowConfig{
						AddOns: []string{
							"example.com/addon",
						},
					},
				},
				sources:       nil,
				errorExpected: true,
			}
		},
		func() *updateAddons {
			url := "example.com/addon"
			m := mocks.MockUpdateSource{}
			m.On("GetURLRegex").Return(regexp.MustCompile("example.com/.+"))
			m.On("GetLatestVersion", url).Return("1.2.3", nil)
			m.On("DownloadAddon", url, "").Return(nil)

			return &updateAddons{
				updater: &gameUpdater{
					config: config.WowConfig{
						AddOns: []string{
							url,
						},
					},
				},
				sources: []UpdateSource{
					&m,
				},
				errorExpected: false,
			}
		},
		func() *updateAddons {
			url := "example.com/addon"
			m := mocks.MockUpdateSource{}
			m.On("GetURLRegex").Return(regexp.MustCompile("example.com/.+"))
			m.On("GetLatestVersion", url).Return("", errors.New("i'm an error"))

			return &updateAddons{
				updater: &gameUpdater{
					config: config.WowConfig{
						AddOns: []string{
							url,
						},
					},
				},
				sources: []UpdateSource{
					&m,
				},
				errorExpected: true,
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		err := tt.updater.updateAddons(tt.sources)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
