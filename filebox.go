package filebox

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/qor/admin"
	"github.com/qor/roles"
)

// Filebox is a based object contains download folder path and admin.Auth used to get current user
type Filebox struct {
	BaseDir string
	Auth    admin.Auth
	prefix  string
}

// File is a object to access a specific file
type File struct {
	FilePath string
	Roles    []string
	Dir      *Dir
	Filebox  *Filebox
}

// Dir is a object to access a specific directory
type Dir struct {
	DirPath string
	Roles   []string
	Filebox *Filebox
}

// ServeHTTP is a implement for http server interface
func (filebox *Filebox) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filebox.Download(w, req)
}

// New a filebox struct
func New(dir string) *Filebox {
	return &Filebox{BaseDir: dir}
}

// MountTo will mount to mux to route `mountto`
func (filebox *Filebox) MountTo(mountTo string, mux *http.ServeMux) {
	filebox.prefix = "/" + strings.Trim(mountTo, "/")
	mux.Handle(filebox.prefix+"/", filebox)
}

// SetAuth will set a admin.Auth struct to Filebox, used to get current user's role
func (filebox *Filebox) SetAuth(auth admin.Auth) {
	filebox.Auth = auth
}

// AccessFile will return a specific File object
func (filebox *Filebox) AccessFile(filePath string, roles ...string) *File {
	file := &File{FilePath: filepath.Join(filebox.BaseDir, filePath), Roles: roles, Filebox: filebox}
	file.Dir = filebox.AccessDir(filepath.Dir(filePath), roles...)
	return file
}

// Read will get a io reader for a specific file
func (f *File) Read() (io.ReadSeeker, error) {
	if f.HasPermission(roles.Read) {
		return os.Open(f.FilePath)
	}
	return nil, roles.ErrPermissionDenied
}

// Write used to store reader's content to a file
func (f *File) Write(reader io.Reader) (err error) {
	if f.HasPermission(roles.Update) {
		var dst *os.File
		if _, err = os.Stat(f.FilePath); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(f.FilePath), os.ModePerm)
		}

		if err == nil {
			if dst, err = os.Create(f.FilePath); err == nil {
				_, err = io.Copy(dst, reader)
			}
		}
		return err
	}
	return roles.ErrPermissionDenied
}

// SetPermission used to set a Permission to file
func (f *File) SetPermission(permission *roles.Permission) (err error) {
	jsonVal, err := json.Marshal(permission)
	if err == nil {
		err = ioutil.WriteFile(f.metaFilePath(), jsonVal, 0644)
	}
	return err
}

// HasPermission used to check current user whether have permission to access file
func (f *File) HasPermission(mode roles.PermissionMode) bool {
	if _, err := os.Stat(f.metaFilePath()); !os.IsNotExist(err) {
		return hasPermission(f.metaFilePath(), mode, f.Roles)
	}
	return f.Dir.HasPermission(mode)
}

func (f *File) metaFilePath() string {
	fileName := filepath.Base(f.FilePath)
	dir := filepath.Dir(f.FilePath)
	return filepath.Join(dir, fileName+".meta")
}

// AccessDir will return a specific Dir object
func (filebox *Filebox) AccessDir(dirPath string, roles ...string) *Dir {
	return &Dir{DirPath: filepath.Join(filebox.BaseDir, dirPath), Roles: roles, Filebox: filebox}
}

// WriteFile writes data to a file named by filename. If the file does not exist, WriteFile will create a new file
func (dir *Dir) WriteFile(fileName string, reader io.Reader) (file *File, err error) {
	if err = dir.createIfNoExist(); err == nil {
		relativeDir := strings.TrimPrefix(dir.DirPath, dir.Filebox.BaseDir)
		file = dir.Filebox.AccessFile(filepath.Join(relativeDir, fileName), dir.Roles...)
		err = file.Write(reader)
	}
	return
}

// SetPermission used to set a Permission to directory
func (dir *Dir) SetPermission(permission *roles.Permission) (err error) {
	err = dir.createIfNoExist()
	jsonVal, err := json.Marshal(permission)
	if err == nil {
		err = ioutil.WriteFile(dir.metaDirPath(), jsonVal, 0644)
	}
	return err
}

// HasPermission used to check current user whether have permission to access directory
func (dir *Dir) HasPermission(mode roles.PermissionMode) bool {
	return hasPermission(dir.metaDirPath(), mode, dir.Roles)
}

func (dir *Dir) createIfNoExist() (err error) {
	if _, err = os.Stat(dir.DirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dir.DirPath, os.ModePerm)
	}
	return err
}

func (dir *Dir) metaDirPath() string {
	return filepath.Join(dir.DirPath, ".meta")
}

func hasPermission(metaFilePath string, mode roles.PermissionMode, currentRoles []string) bool {
	if _, err := os.Stat(metaFilePath); !os.IsNotExist(err) {
		if bytes, err := ioutil.ReadFile(metaFilePath); err == nil {
			permission := &roles.Permission{}
			if json.Unmarshal(bytes, permission); err == nil {
				for _, role := range currentRoles {
					if permission.HasPermission(mode, role) {
						return true
					}
				}
			}
		}
		return false
	}
	return true
}
