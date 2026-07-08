package infra

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func downloadFile(url, path string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("download %s failed: %s", url, resp.Status)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func replaceSignedBy(input string) string {
	return strings.ReplaceAll(input, "deb https://", "deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://")
}
