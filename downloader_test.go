package downloader_test

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/qor/downloader"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloader(t *testing.T) {
	root, _ := os.Getwd()
	downloader := downloader.New(root + "/test/download")
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	downloader.MountTo(mux)
	req, err := http.Get(server.URL + "/download/a.csv")
	if err != nil || req.StatusCode != 200 {
		t.Errorf(color.RedString(fmt.Sprintf("Downloader error: can't get file")))
	}
	body, _ := ioutil.ReadAll(req.Body)
	if string(body) != "Column1,Column2\n" {
		t.Errorf(color.RedString(fmt.Sprintf("Downloader error: file'content is incorrect")))
	}
}
