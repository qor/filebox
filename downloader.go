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

type Downloader struct {
	Prefix   string
	FilePath string
	Auth     admin.Auth
}

func (downloader *Downloader) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	downloader.Download(w, req)
}

func (downloader *Downloader) MountTo(mux *http.ServeMux) {
	mux.Handle("/downloads/", downloader)
}

func New(prefix string) *Downloader {
	return &Downloader{
		Prefix: prefix,
	}
}

func (downloader *Downloader) SetAuth(auth admin.Auth) {
	downloader.Auth = auth
}

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

func (downloader *Downloader) Get(filePath string) *Downloader {
	return &Downloader{Prefix: downloader.Prefix, FilePath: filePath}
}

func (downloader *Downloader) SetPermission(permission *roles.Permission) error {
	jsonVal, err := json.Marshal(permission)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fullMetaFilePath(downloader.fullFilePath()), jsonVal, 0644)
	return err
}

func (downloader *Downloader) fullFilePath() string {
	return path.Join(downloader.Prefix, downloader.FilePath)
}

func fullMetaFilePath(fullFilePath string) string {
	fileName := filepath.Base(fullFilePath)
	dir := filepath.Dir(fullFilePath)
	return path.Join(dir, fileName+".meta")
}
