package filebox

import (
	"encoding/json"
	"github.com/qor/admin"
	"github.com/qor/roles"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// Filebox is a object save download folder path and a specific download file used to set permission
type Filebox struct {
	Dir      string
	FilePath string
	Auth     admin.Auth
}

func (filebox *Filebox) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filebox.Download(w, req)
}

// New a filebox struct with download dir
func New(dir string) *Filebox {
	return &Filebox{Dir: dir}
}

// MountTo will mount `/downloads` to mux
func (filebox *Filebox) MountTo(mux *http.ServeMux) {
	mux.Handle("/downloads/", filebox)
}

// SetAuth will set a admin.Auth struct to Filebox, used to get current user's role
func (filebox *Filebox) SetAuth(auth admin.Auth) {
	filebox.Auth = auth
}

// Get will return a new Filebox with a specific file
func (filebox *Filebox) Get(filePath string) *Filebox {
	return &Filebox{Dir: filebox.Dir, FilePath: filePath}
}

// Put will read context from reader and save as file then return a new Filebox with this new file
func (filebox *Filebox) Put(filePath string, reader io.Reader) (newFilebox *Filebox, err error) {
	newFilebox = filebox.Get(filePath)
	var fullFilePath = newFilebox.fullFilePath()
	var dst *os.File
	if _, err = os.Stat(filepath.Dir(fullFilePath)); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(fullFilePath), os.ModePerm)
	}
	if dst, err = os.Create(newFilebox.fullFilePath()); err == nil {
		_, err = io.Copy(dst, reader)
	}
	return newFilebox, err
}

// SetPermission will set a permission to file used to control access
func (filebox *Filebox) SetPermission(permission *roles.Permission) error {
	jsonVal, err := json.Marshal(permission)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fullMetaFilePath(filebox.fullFilePath()), jsonVal, 0644)
	return err
}

func (filebox *Filebox) fullFilePath() string {
	return path.Join(filebox.Dir, filebox.FilePath)
}

func fullMetaFilePath(fullFilePath string) string {
	fileName := filepath.Base(fullFilePath)
	dir := filepath.Dir(fullFilePath)
	return path.Join(dir, fileName+".meta")
}
