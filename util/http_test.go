package util

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckHTTPResponse(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		resp    *http.Response
		wantErr bool
	}{
		{
			name: "random error",
			err:  errors.New("i am an error"),
			resp: &http.Response{
				StatusCode: 200,
			},
			wantErr: true,
		},
		{
			name:    "random error, nil resp",
			err:     errors.New("i am an error"),
			resp:    nil,
			wantErr: true,
		},
		{
			name: "status code 200",
			err:  nil,
			resp: &http.Response{
				StatusCode: 200,
			},
			wantErr: false,
		},
		{
			name: "status code 299",
			err:  nil,
			resp: &http.Response{
				StatusCode: 299,
			},
			wantErr: false,
		},
		{
			name: "status code 300",
			err:  nil,
			resp: &http.Response{
				StatusCode: 300,
			},
			wantErr: true,
		},
		{
			name: "status code 199",
			err:  nil,
			resp: &http.Response{
				StatusCode: 199,
			},
			wantErr: true,
		},
		{
			name:    "empty response",
			err:     nil,
			resp:    &http.Response{},
			wantErr: true,
		},
		{
			name:    "nil response",
			err:     nil,
			resp:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckHTTPResponse(tt.resp, tt.err)

			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestGetHTMLPage(t *testing.T) {
	t.Run("failed http call", func(t *testing.T) {
		_, err := GetHTMLPage(new(http.Client), "no url")

		assert.Error(t, err)
	})
	t.Run("empty document", func(t *testing.T) {
		m := http.NewServeMux()
		m.HandleFunc("/", func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(http.StatusNoContent)
		})
		s := httptest.NewServer(m)
		defer s.Close()

		doc, err := GetHTMLPage(new(http.Client), s.URL)

		assert.NoError(t, err)
		assert.NotNil(t, doc)
	})
	t.Run("example document", func(t *testing.T) {
		m := http.NewServeMux()
		m.HandleFunc("/", func(rw http.ResponseWriter, _ *http.Request) {
			_, _ = rw.Write([]byte(`
<!DOCTYPE html>
<html lang="en-US">
	<head>
		<title>Tukui</title>
	</head>
	<body class="appear-animate">
	<div class="tab-pane fade in" id="extras">
		<p class="extras">The latest version of this addon is <b class="VIP">%s</b> and was uploaded on <b class="VIP">Oct 27, 2020</b> at <b class="VIP">02:17</b>.</p>
		<p class="extras">This file was last downloaded on <b class="VIP">Dec 09, 2020</b> at <b class="VIP">21:48</b> and has been downloaded <b class="VIP">1572354</b> times.</p>
	</div>
	</body>
</html>
	`))
		})
		s := httptest.NewServer(m)
		defer s.Close()

		doc, err := GetHTMLPage(new(http.Client), s.URL)

		assert.NoError(t, err)
		assert.NotNil(t, doc)
	})
}
