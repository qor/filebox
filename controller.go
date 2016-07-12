package filebox

import (
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Download is a handler will return a specific file
func (filebox *Filebox) Download(w http.ResponseWriter, req *http.Request) {
	filePath := strings.Replace(req.URL.Path, filebox.prefix, "", 1)
	context := &admin.Context{Context: &qor.Context{Request: req, Writer: w}}
	allRoles := roles.MatchedRoles(req, filebox.Auth.GetCurrentUser(context))
	file := filebox.AccessFile(filePath, allRoles...)
	if reader, err := file.Read(); err == nil {
		fileName := filepath.Base(file.FilePath)
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Header().Set("Content-Type", req.Header.Get("Content-Type"))
		http.ServeContent(w, req, fileName, time.Now(), reader)
		return
	}
	http.NotFound(w, req)
}
