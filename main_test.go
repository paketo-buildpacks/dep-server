package main_test

import (
	"crypto/tls"
	"fmt"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	spec.Run(t, "Server", testServer, spec.Report(report.Terminal{}))
}

func testServer(t *testing.T, when spec.G, it spec.S) {
	var (
		serverPath   string
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

		tempFile, err := ioutil.TempFile("", "dep-server")
		require.NoError(err)

		serverPath = tempFile.Name()
		require.NoError(tempFile.Close())

		goBuild := exec.Command("go", "build", "-o", serverPath, ".")
		output, err := goBuild.CombinedOutput()
		require.NoError(err, "failed to build server: %s", string(output))
	})

	it.After(func() {
		testS3Server.Close()
		_ = os.Remove(serverPath)
	})

	when("/api/v1/metadata", func() {
		it("returns the metadata file for the given dependency", func() {
			addr, err := GetFreeAddr()
			require.NoError(err)

			go func() {
				output, err := exec.Command(
					serverPath,
					"--addr", addr,
					"--s3-url", testS3Server.URL,
				).CombinedOutput()
				require.NoError(err, string(output))
			}()
			err = WaitForServerToBeAvailable(addr, 10*time.Second)
			require.NoError(err)

			resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/dependency?name=some-dep", addr))
			require.NoError(err)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(err)

			expectedOutput := `[{"name": "some-dep","version": "2.0.0"}, {"name": "some-dep","version": "1.0.0"}]`
			assert.JSONEq(expectedOutput, string(body))
		})
	})
}

func GetFreeAddr() (string, error) {
	conn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return conn.Addr().String(), nil
}

func ServerIsAvailable(address string) bool {
	conn, err := net.Dial("tcp", address)
	if err == nil {
		_ = tls.Client(conn, &tls.Config{InsecureSkipVerify: true}).Handshake()
		_ = conn.Close()

		return true
	}

	return false
}

func WaitForServerToBeAvailable(address string, timeout time.Duration) error {
	timeoutChan := time.After(timeout)

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("failed to connect to %s within %s", address, timeout)
		default:
			if ServerIsAvailable(address) {
				return nil
			}
		}
	}
}
