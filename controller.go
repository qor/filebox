package downloader

import (
	"encoding/json"
	"github.com/qor/roles"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

func (downloader *Downloader) Download(w http.ResponseWriter, req *http.Request) {
	filePath := strings.Replace(req.URL.Path, "/downloads", "", 1)
	fullFilePath := path.Join(downloader.Prefix, filePath)
	if _, err := os.Stat(fullFilePath); os.IsNotExist(err) {
		http.NotFound(w, req)
	} else {
		if hasPermission(fullFilePath, "user") {
			http.ServeFile(w, req, fullFilePath)
			return
		}
		http.NotFound(w, req)
	}
}

func hasPermission(fullFilePath string, role string) bool {
	if _, err := os.Stat(fullMetaFilePath(fullFilePath)); !os.IsNotExist(err) {
		bytes, err := ioutil.ReadFile(fullMetaFilePath(fullFilePath))
		if err != nil {
			return false
		}
		permission := &roles.Permission{}
		err = json.Unmarshal(bytes, permission)
		if err == nil {
			return permission.HasPermission(roles.Read, role)
		}
	}
	return true
}
