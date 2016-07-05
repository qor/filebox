package downloader_test

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/qor/admin"
	"github.com/qor/downloader"
	"github.com/qor/qor"
	"github.com/qor/roles"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var Downloader *downloader.Downloader
var Admin *admin.Admin
var Server *httptest.Server
var user *User

// User definition
type User struct {
	Name string
	Role string
}

func (user *User) DisplayName() string {
	return user.Name
}

type AdminAuth struct {
}

func (AdminAuth) LoginURL(c *admin.Context) string {
	return "/auth/login"
}

func (AdminAuth) LogoutURL(c *admin.Context) string {
	return "/auth/logout"
}

func (AdminAuth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	return user
}

// Init
func init() {
	root, _ := os.Getwd()
	os.Remove(root + "/test/downloads/a.csv.meta")
	mux := http.NewServeMux()
	Server = httptest.NewServer(mux)

	user = &User{Name: "user", Role: "normal_user"}
	roles.Register("admin", func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*User).Role == "admin"
	})

	Downloader = downloader.New(root + "/test/downloads")
	Downloader.MountTo(mux)
	Downloader.SetAuth(AdminAuth{})
}

// Test download cases
func TestDownloader(t *testing.T) {
	req, err := http.Get(Server.URL + "/downloads/a.csv")
	if err != nil || req.StatusCode != 200 {
		t.Errorf(color.RedString(fmt.Sprintf("Downloader error: can't get file")))
	}
	body, _ := ioutil.ReadAll(req.Body)
	if string(body) != "Column1,Column2\n" {
		t.Errorf(color.RedString(fmt.Sprintf("Downloader error: file'content is incorrect")))
	}

	permission := roles.Allow(roles.Read, "admin")
	if err := Downloader.Get("a.csv").SetPermission(permission); err != nil {
		t.Errorf(color.RedString(fmt.Sprintf("Downloader error: create meta file failure (%v)", err)))
	}

	req, err = http.Get(Server.URL + "/downloads/a.csv")
	if err != nil || req.StatusCode != 404 {
		t.Errorf(color.RedString(fmt.Sprintf("Downloader error: should can't download file")))
	}
}
