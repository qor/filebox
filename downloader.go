package downloader

import (
	"net/http"
)

type Downloader struct {
	Prefix string
}

func (downloader *Downloader) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	Download(w, req)
}

func (downloader *Downloader) MountTo(mux *http.ServeMux) {
	mux.Handle("/download", downloader)
}

func New(prefix string) *Downloader {
	return &Downloader{
		Prefix: prefix,
	}
}

func (downloader *Downloader) Put(filePath string) *Downloader {
	return downloader
}

func (downloader *Downloader) Get(filePath string) *Downloader {
	return downloader
}

func (downloader *Downloader) SetPermission() {
}
