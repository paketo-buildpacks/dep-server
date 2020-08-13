package handler_test

import (
	"fmt"
	h "github.com/pivotal/dep-server/handler"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	spec.Run(t, "Handler", testHandler, spec.Report(report.Terminal{}))
}

func testHandler(t *testing.T, when spec.G, it spec.S) {
	var (
		handler      h.Handler
		testS3Server *httptest.Server
		assert       = assert.New(t)
		require      = require.New(t)
	)

	it.Before(func() {
		testS3Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() == "/pivotal-buildpacks/metadata/some-dep.json" {
				_, _ = fmt.Fprintln(w, `[{"name": "some-dep","version": "2.0.0"}, {"name": "some-dep","version": "1.0.0"}]`)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		handler = h.Handler{S3URL: testS3Server.URL}
	})

	it("returns the contents of the file in s3", func() {
		req := httptest.NewRequest("GET", "http://some-url.com/some-endpoint?name=some-dep", nil)
		w := httptest.NewRecorder()
		handler.DependencyHandler(w, req)

		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(err)

		assert.JSONEq(`[{"name": "some-dep","version": "2.0.0"}, {"name": "some-dep","version": "1.0.0"}]`, string(body))
	})

	when("the request is not a GET", func() {
		it("returns a 405", func() {
			req := httptest.NewRequest("POST", "http://some-url.com/some-endpoint?name=some-dep", nil)
			w := httptest.NewRecorder()
			handler.DependencyHandler(w, req)

			resp := w.Result()
			assert.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
		})
	})

	when("the request does not include a dependency name", func() {
		it("returns a 400", func() {
			req := httptest.NewRequest("GET", "http://some-url.com/some-endpoint", nil)
			w := httptest.NewRecorder()
			handler.DependencyHandler(w, req)

			resp := w.Result()
			assert.Equal(http.StatusBadRequest, resp.StatusCode)
		})
	})

	when("the s3 server responds with a non-200", func() {
		it("returns a 500", func() {
			req := httptest.NewRequest("GET", "http://some-url.com/some-endpoint?name=some-non-existent-dep", nil)
			w := httptest.NewRecorder()
			handler.DependencyHandler(w, req)

			resp := w.Result()
			assert.Equal(http.StatusInternalServerError, resp.StatusCode)
		})
	})
}
