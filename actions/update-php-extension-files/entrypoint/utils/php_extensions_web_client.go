package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . PHPExtensionsWebClient
type PHPExtensionsWebClient interface {
	DownloadExtensionsSource(url, filename string) error
}

type WebClient struct {
	client *http.Client
}

func NewPHPExtensionsWebClient() WebClient {
	return WebClient{
		client: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

func (w WebClient) DownloadExtensionsSource(url, filename string) error {
	response, err := w.client.Get(url)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("got unsuccessful response: status code: %d, body: %s", response.StatusCode, body)
	}

	defer response.Body.Close()

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
