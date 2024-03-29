// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	regexp "regexp"

	mock "github.com/stretchr/testify/mock"
)

// MockUpdateSource is an autogenerated mock type for the UpdateSource type
type MockUpdateSource struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockUpdateSource) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DownloadAddon provides a mock function with given fields: addonURL, dir
func (_m *MockUpdateSource) DownloadAddon(addonURL string, dir string) error {
	ret := _m.Called(addonURL, dir)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(addonURL, dir)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetLatestVersion provides a mock function with given fields: addonURL
func (_m *MockUpdateSource) GetLatestVersion(addonURL string) (string, error) {
	ret := _m.Called(addonURL)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(addonURL)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(addonURL)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetURLRegex provides a mock function with given fields:
func (_m *MockUpdateSource) GetURLRegex() *regexp.Regexp {
	ret := _m.Called()

	var r0 *regexp.Regexp
	if rf, ok := ret.Get(0).(func() *regexp.Regexp); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*regexp.Regexp)
		}
	}

	return r0
}
