package filebox

import (
	"encoding/json"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

// Download is a handler will return a specific file
func (filebox *Filebox) Download(w http.ResponseWriter, req *http.Request) {
	filePath := strings.Replace(req.URL.Path, "/downloads", "", 1)
	fullFilePath := path.Join(filebox.Dir, filePath)
	if _, err := os.Stat(fullFilePath); os.IsNotExist(err) {
		http.NotFound(w, req)
	} else {
		if filebox.hasPermission(fullFilePath, w, req) {
			http.ServeFile(w, req, fullFilePath)
			return
		}
		http.NotFound(w, req)
	}
}

func (filebox *Filebox) hasPermission(fullFilePath string, w http.ResponseWriter, req *http.Request) bool {
	if _, err := os.Stat(fullMetaFilePath(fullFilePath)); !os.IsNotExist(err) {
		bytes, err := ioutil.ReadFile(fullMetaFilePath(fullFilePath))
		if err != nil {
			return false
		}
		permission := &roles.Permission{}
		err = json.Unmarshal(bytes, permission)
		if err == nil {
			context := &admin.Context{Context: &qor.Context{Request: req, Writer: w}}
			allRoles := roles.MatchedRoles(req, filebox.Auth.GetCurrentUser(context))
			var hasPermission bool
			for _, role := range allRoles {
				if permission.HasPermission(roles.Read, role) {
					hasPermission = true
					break
				}
			}
			return hasPermission
		}
	}
	return true
}
