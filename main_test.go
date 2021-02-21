package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unly/wow-addon-updater/updater"
	"github.com/unly/wow-addon-updater/updater/mocks"
	"github.com/unly/wow-addon-updater/util"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func Test_generateDefaultConfig(t *testing.T) {
	type configTest struct {
		path          string
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *configTest{
		func() *configTest {
			dir := helpers.TempDir(t)

			return &configTest{
				path:          filepath.Join(dir, "config"),
				errorExpected: false,
				teardown:      helpers.DeleteDir(t, dir),
			}
		},
		func() *configTest {
			return &configTest{
				path:          "",
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *configTest {
			return &configTest{
				path:          ".",
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		err := generateDefaultConfig(tt.path)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.FileExists(t, tt.path)
		}

		tt.teardown()
	}
}

func Test_closeSources(t *testing.T) {
	m1 := new(mocks.MockUpdateSource)
	m1.On("Close")
	m2 := new(mocks.MockUpdateSource)
	m2.On("Close")
	addonSources = []updater.UpdateSource{
		m1,
		m2,
	}

	closeSources(addonSources)

	assert.Equal(t, 1, len(m1.Calls))
	assert.Equal(t, "Close", m1.Calls[0].Method)
	assert.Equal(t, 1, len(m2.Calls))
	assert.Equal(t, "Close", m2.Calls[0].Method)
}

func Test_runAndRecover(t *testing.T) {
	type mainTest struct {
		args          []string
		errorExpected bool
		checks        func()
		teardown      helpers.TearDown
	}
	oldSources := addonSources
	oldArgs := os.Args
	defer func() {
		addonSources = oldSources
		os.Args = oldArgs
	}()
	addonSources = []updater.UpdateSource{}

	tests := []func() *mainTest{
		func() *mainTest {
			return &mainTest{
				args:          []string{"-c"},
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *mainTest {
			return &mainTest{
				args:          []string{"-c", "."},
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *mainTest {
			return &mainTest{
				args:          []string{"-c", "."},
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *mainTest {
			file := "file"

			return &mainTest{
				args:          []string{"-c", file},
				errorExpected: false,
				checks: func() {
					assert.True(t, util.FileExists(file))
				},
				teardown: helpers.DeleteFile(t, file),
			}
		},
		func() *mainTest {
			content := []byte(`
classic:
  path: path/to/classic
  addons:
    - addon1
    - addon2`)
			file := helpers.TempFile(t, "", content)
			err := os.WriteFile(file, content, os.FileMode(0666))
			assert.NoError(t, err)

			return &mainTest{
				args:          []string{"-c", file},
				errorExpected: true,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
		func() *mainTest {
			m := new(mocks.MockUpdateSource)
			m.On("GetURLRegex").Return(regexp.MustCompile(`addon.+`))
			m.On("GetLatestVersion", mock.Anything).Return("1.2.3", nil)
			m.On("DownloadAddon", mock.Anything, mock.Anything).Return(nil)
			m.On("Close")
			addonSources = []updater.UpdateSource{
				m,
			}

			content := []byte(`
classic:
  path: path/to/classic
  addons:
    - addon1
    - addon2`)
			dir := helpers.TempDir(t)
			file := helpers.TempFile(t, dir, content)
			err := os.WriteFile(file, content, os.FileMode(0666))
			assert.NoError(t, err)
			oldVersionsPath := versionsPath
			versionsPath = filepath.Join(dir, ".versions")

			return &mainTest{
				args:          []string{"-c", file},
				errorExpected: false,
				checks: func() {
					assert.Equal(t, 7, len(m.Calls))
				},
				teardown: func() {
					helpers.DeleteDir(t, dir)
					addonSources = []updater.UpdateSource{}
					versionsPath = oldVersionsPath
				},
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		os.Args = []string{
			oldArgs[0],
		}
		os.Args = append(os.Args, tt.args...)
		err := runAndRecover()

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		if tt.checks != nil {
			tt.checks()
		}

		tt.teardown()
	}
}
