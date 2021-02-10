package internal_test

import (
	"fmt"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWebClient(t *testing.T) {
	spec.Run(t, "webClient", testWebClient, spec.Report(report.Terminal{}))
}

func testWebClient(t *testing.T, when spec.G, it spec.S) {
	var (
		assert    = assert.New(t)
		require   = require.New(t)
		server    *httptest.Server
		webClient internal.WebClient
		testDir   string
	)
	const (
		fileContents = "some-contents"
	)

	it.Before(func() {
		var err error
		testDir, err = ioutil.TempDir("", "external-dependency-resource-web-client")
		require.NoError(err)

		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/file":
				_, _ = fmt.Fprint(w, fileContents)
			case "/headers":
				_, _ = fmt.Fprint(w, r.Header)
			case "/body":
				defer r.Body.Close()
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(err)
				_, _ = fmt.Fprint(w, string(body))
			case "/500":
				w.WriteHeader(500)
				_, _ = fmt.Fprint(w, "some-server-error")
			}
		}))

		webClient = internal.NewWebClient()
	})

	it.After(func() {
		_ = os.RemoveAll(testDir)
		server.Close()
	})

	when("Download", func() {
		it("downloads the file", func() {
			outputPath := filepath.Join(testDir, "some-file.txt")

			err := webClient.Download(server.URL+"/file", outputPath)
			require.NoError(err)

			contents, err := ioutil.ReadFile(outputPath)
			require.NoError(err)
			assert.Equal(fileContents, string(contents))
		})

		when("the response is not a 200", func() {
			it("returns an error", func() {
				err := webClient.Download(server.URL+"/500", "")
				assert.Error(err)
				assert.Equal("got unsuccessful response: status code: 500, body: some-server-error", err.Error())
			})
		})
	})

	when("Get", func() {
		it("returns the request body", func() {
			responseBody, err := webClient.Get(server.URL + "/file")
			require.NoError(err)

			assert.Equal(fileContents, string(responseBody))
		})

		when("WithHeader is specified", func() {
			it("adds the header to the request", func() {
				responseBody, err := webClient.Get(server.URL+"/headers", internal.WithHeader("some-key", "some-value"))
				require.NoError(err)

				assert.Contains(string(responseBody), "Some-Key:[some-value]")
			})
		})

		when("the response is not a 200", func() {
			it("returns an error", func() {
				_, err := webClient.Get(server.URL + "/500")
				assert.Error(err)
				assert.Equal("got unsuccessful response: status code: 500, body: some-server-error", err.Error())
			})
		})
	})

	when("Post", func() {
		it("returns the request body", func() {
			responseBody, err := webClient.Post(server.URL+"/body", []byte("some-request-body"))
			require.NoError(err)

			assert.Equal("some-request-body", string(responseBody))
		})

		when("WithHeader is specified", func() {
			it("adds the header to the request", func() {
				responseBody, err := webClient.Post(server.URL+"/headers", nil, internal.WithHeader("some-key", "some-value"))
				require.NoError(err)

				assert.Contains(string(responseBody), "Some-Key:[some-value]")
			})
		})

		when("the response is not a 200", func() {
			it("returns an error", func() {
				_, err := webClient.Post(server.URL+"/500", nil)
				assert.Error(err)
				assert.Equal("got unsuccessful response: status code: 500, body: some-server-error", err.Error())
			})
		})
	})
}
