package filebox

import (
	"encoding/json"
	"fmt"
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

type File struct {
	FilePath string
	Roles    []string
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

func (filebox *Filebox) AccessFile(filePath string, roles ...string) *File {
	return &File{FilePath: path.Join(filebox.Dir, filePath), Roles: roles}
}

func (file *File) Read() (string, error) {
	if HasPermission(file.FilePath, file.Roles...) {
		bytes, err := ioutil.ReadFile(file.FilePath)
		return string(bytes), err
	}
	return "", fmt.Errorf("Doesn't have permission to read the file")
}

func (file *File) Write(reader io.Reader) (err error) {
	var dst *os.File
	if _, err = os.Stat(file.FilePath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(file.FilePath), os.ModePerm)
	}
	if err != nil {
		return err
	}
	if dst, err = os.Create(file.FilePath); err == nil {
		_, err = io.Copy(dst, reader)
	}
	return err
}

func (file *File) SetPermission(permission *roles.Permission) (err error) {
	jsonVal, err := json.Marshal(permission)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file.metaFilePath(), jsonVal, 0644)
	return err
}

func (file *File) metaFilePath() string {
	fileName := filepath.Base(file.FilePath)
	dir := filepath.Dir(file.FilePath)
	return path.Join(dir, fileName+".meta")
}

func HasPermission(fullFilePath string, currentRoles ...string) bool {
	if _, err := os.Stat(fullMetaFilePath(fullFilePath)); !os.IsNotExist(err) {
		bytes, err := ioutil.ReadFile(fullMetaFilePath(fullFilePath))
		if err != nil {
			return false
		}
		permission := &roles.Permission{}
		err = json.Unmarshal(bytes, permission)
		if err == nil {
			var hasPermission bool
			for _, role := range currentRoles {
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

func (filebox *Filebox) fullFilePath() string {
	return path.Join(filebox.Dir, filebox.FilePath)
}

func fullMetaFilePath(fullFilePath string) string {
	fileName := filepath.Base(fullFilePath)
	dir := filepath.Dir(fullFilePath)
	return path.Join(dir, fileName+".meta")
}
