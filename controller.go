package downloader

import (
	"net/http"
	"os"
	"path"
	"strings"
)

func (downloader *Downloader) Download(w http.ResponseWriter, req *http.Request) {
	filePath := strings.Replace(req.URL.Path, "/download", "", 1)
	fullFilePath := path.Join(downloader.Prefix, filePath)
	if _, err := os.Stat(fullFilePath); os.IsNotExist(err) {
		http.NotFound(w, req)
	} else {
		http.ServeFile(w, req, fullFilePath)
	}
}
