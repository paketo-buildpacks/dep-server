package main_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	spec.Run(t, "Server", testServer, spec.Report(report.Terminal{}))
}

func testServer(t *testing.T, when spec.G, it spec.S) {
	var (
		serverPath       string
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

		tempFile, err := ioutil.TempFile("", "dep-server")
		require.NoError(err)

		serverPath = tempFile.Name()
		require.NoError(tempFile.Close())

		goBuild := exec.Command("go", "build", "-o", serverPath, ".")
		output, err := goBuild.CombinedOutput()
		require.NoError(err, "failed to build server: %s", string(output))
	})

	it.After(func() {
		testBucketServer.Close()
		_ = os.Remove(serverPath)
	})

	when("/v1/metadata", func() {
		it("returns the metadata file for the given dependency", func() {
			port, err := GetFreePort()
			require.NoError(err)

			go func() {
				cmd := exec.Command(serverPath, "--bucket-url", testBucketServer.URL)
				cmd.Env = append(cmd.Env, "PORT="+port)
				output, err := cmd.CombinedOutput()
				require.NoError(err, string(output))
			}()
			err = WaitForServerToBeAvailable(port, 10*time.Second)
			require.NoError(err)

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/v1/dependency?name=some-dep", port))
			require.NoError(err)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(err)

			expectedOutput := `[{"name": "some-dep","version": "2.0.0"}, {"name": "some-dep","version": "1.0.0"}]`
			assert.JSONEq(expectedOutput, string(body))
		})
	})
}

func GetFreePort() (string, error) {
	conn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return strings.Split(conn.Addr().String(), ":")[1], nil
}

func ServerIsAvailable(port string) bool {
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
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
