package handler_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/paketo-buildpacks/dep-server/internal/handler"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	spec.Run(t, "Handler", testHandler, spec.Report(report.Terminal{}))
}

func testHandler(t *testing.T, when spec.G, it spec.S) {
	var (
		handler          h.Handler
		testBucketServer *httptest.Server
		assert           = assert.New(t)
		require          = require.New(t)
	)

	it.Before(func() {
		testBucketServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() == "/metadata/some-dep.json" {
				_, _ = fmt.Fprintln(w, `[{"name": "some-dep","version": "2.0.0"}, {"name": "some-dep","version": "1.0.0"}]`)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		handler = h.Handler{BucketURL: testBucketServer.URL}
	})

	it("returns the contents of the file in the bucket", func() {
		req := httptest.NewRequest("GET", "http://some-url.com/some-endpoint?name=some-dep", nil)
		w := httptest.NewRecorder()
		handler.DependencyHandler(w, req)

		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(err)

		assert.JSONEq(`[{"name": "some-dep","version": "2.0.0"}, {"name": "some-dep","version": "1.0.0"}]`, string(body))
	})

	it("converts all dep-names to lowercase before making a request to the bucket", func() {
		req := httptest.NewRequest("GET", "http://some-url.com/some-endpoint?name=some-DEP", nil)
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

	when("the bucket server responds with a non-200", func() {
		it("returns a 500", func() {
			req := httptest.NewRequest("GET", "http://some-url.com/some-endpoint?name=some-non-existent-dep", nil)
			w := httptest.NewRecorder()
			handler.DependencyHandler(w, req)

			resp := w.Result()
			assert.Equal(http.StatusInternalServerError, resp.StatusCode)
		})
	})
}
