package sources

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func Test_Close(t *testing.T) {
	tests := []struct {
		source *source
	}{
		{
			source: &source{
				tempDir: helpers.TempDir(t),
			},
		},
		{
			source: &source{
				tempDir: "hello world",
			},
		},
	}

	for _, tt := range tests {
		tt.source.Close()

		assert.NoDirExists(t, tt.source.tempDir)
	}
}

func Test_checkHTTPResponse(t *testing.T) {
	tests := []struct {
		response      *http.Response
		err           error
		errorExpected bool
	}{
		{
			response:      nil,
			err:           nil,
			errorExpected: false,
		},
		{
			response:      nil,
			err:           errors.New("I am an error"),
			errorExpected: true,
		},
		{
			response: &http.Response{
				StatusCode: http.StatusOK,
			},
			err:           nil,
			errorExpected: false,
		},
		{
			response: &http.Response{
				StatusCode: http.StatusNoContent,
			},
			err:           nil,
			errorExpected: false,
		},
		{
			response: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			err:           nil,
			errorExpected: false,
		},
		{
			response: &http.Response{
				StatusCode: http.StatusCreated,
			},
			err:           nil,
			errorExpected: false,
		},
		{
			response: &http.Response{
				StatusCode: http.StatusOK,
			},
			err:           errors.New("I am an error"),
			errorExpected: true,
		},
		{
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
			},
			err:           nil,
			errorExpected: true,
		},
		{
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			err:           nil,
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		actual := checkHTTPResponse(tt.response, tt.err)

		assert.Equal(t, tt.errorExpected, actual != nil)
	}
}
