package filebox

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
)

// Download is a handler will return a specific file
func (filebox *Filebox) Download(w http.ResponseWriter, req *http.Request) {
	var (
		currentUser qor.CurrentUser
		filePath    = strings.TrimPrefix(req.URL.Path, filebox.prefix)
		context     = &admin.Context{Context: &qor.Context{Request: req, Writer: w}}
	)

	if auth := filebox.Auth; auth != nil {
		currentUser = auth.GetCurrentUser(context)
	}

	matchedRoles := roles.MatchedRoles(req, currentUser)

	file := filebox.AccessFile(filePath, matchedRoles...)
	if reader, err := file.Read(); err == nil {
		fileName := filepath.Base(file.FilePath)

		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Header().Set("Content-Type", req.Header.Get("Content-Type"))
		http.ServeContent(w, req, fileName, time.Now(), reader)
	} else if err == roles.ErrPermissionDenied && filebox.Auth != nil {
		http.Redirect(w, req, filebox.Auth.LoginURL(context), http.StatusFound)
	} else {
		http.NotFound(w, req)
	}
}
