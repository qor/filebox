package downloader

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

// Downloader is a object save download folder path and a specific download file used to set permission
type Downloader struct {
	Dir      string
	FilePath string
	Auth     admin.Auth
}

func (downloader *Downloader) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	downloader.Download(w, req)
}

// New a downloader struct with download dir
func New(dir string) *Downloader {
	return &Downloader{Dir: dir}
}

// MountTo will mount `/downloads` to mux
func (downloader *Downloader) MountTo(mux *http.ServeMux) {
	mux.Handle("/downloads/", downloader)
}

// SetAuth will set a admin.Auth struct to Downloader, used to get current user's role
func (downloader *Downloader) SetAuth(auth admin.Auth) {
	downloader.Auth = auth
}

// Get will return a new Downloader with a specific file
func (downloader *Downloader) Get(filePath string) *Downloader {
	return &Downloader{Dir: downloader.Dir, FilePath: filePath}
}

// Put will read context from reader and save as file then return a new Downloader with this new file
func (downloader *Downloader) Put(filePath string, reader io.Reader) (newDownloader *Downloader, err error) {
	newDownloader = downloader.Get(filePath)
	var fullFilePath = newDownloader.fullFilePath()
	var dst *os.File
	if _, err = os.Stat(filepath.Dir(fullFilePath)); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(fullFilePath), os.ModePerm)
	}
	if dst, err = os.Create(newDownloader.fullFilePath()); err == nil {
		_, err = io.Copy(dst, reader)
	}
	return newDownloader, err
}

// SetPermission will set a permission to file used to control access
func (downloader *Downloader) SetPermission(permission *roles.Permission) error {
	jsonVal, err := json.Marshal(permission)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fullMetaFilePath(downloader.fullFilePath()), jsonVal, 0644)
	return err
}

func (downloader *Downloader) fullFilePath() string {
	return path.Join(downloader.Dir, downloader.FilePath)
}

func fullMetaFilePath(fullFilePath string) string {
	fileName := filepath.Base(fullFilePath)
	dir := filepath.Dir(fullFilePath)
	return path.Join(dir, fileName+".meta")
}
