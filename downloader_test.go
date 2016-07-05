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
var CurrentUser *User

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
	return CurrentUser
}

// Init
func init() {
	root, _ := os.Getwd()
	clearMetaFiles(root)
	mux := http.NewServeMux()
	Server = httptest.NewServer(mux)
	CurrentUser = &User{Name: "user", Role: "normal_user"}
	roles.Register("admin", func(req *http.Request, currentUser interface{}) bool {
		return currentUser.(*User) != nil && currentUser.(*User).Role == "admin"
	})
	roles.Register("manager", func(req *http.Request, currentUser interface{}) bool {
		return currentUser.(*User) != nil && currentUser.(*User).Role == "manager"
	})

	Downloader = downloader.New(root + "/test/downloads")
	Downloader.MountTo(mux)
	Downloader.SetAuth(AdminAuth{})
}

// Test download cases
type filePermission struct {
	FileName   string
	AllowRoles []string
}

type testDownloadCase struct {
	CurrentRole      string
	DownloadURL      string
	ExpectStatusCode int
	ExpectContext    string
}

func TestDownloads(t *testing.T) {
	filePermissions := []filePermission{
		filePermission{FileName: "a.csv", AllowRoles: []string{}},
		filePermission{FileName: "b.csv", AllowRoles: []string{"admin"}},
		filePermission{FileName: "c.csv", AllowRoles: []string{"manager", "admin"}},
	}

	testCases := []testDownloadCase{
		testDownloadCase{CurrentRole: "", DownloadURL: "/downloads/missing.csv", ExpectStatusCode: 404, ExpectContext: ""},
		testDownloadCase{CurrentRole: "", DownloadURL: "/downloads/a.csv", ExpectStatusCode: 200, ExpectContext: "Column1,Column2\n"},
		testDownloadCase{CurrentRole: "admin", DownloadURL: "/downloads/a.csv", ExpectStatusCode: 200, ExpectContext: "Column1,Column2\n"},
		testDownloadCase{CurrentRole: "", DownloadURL: "/downloads/b.csv", ExpectStatusCode: 404, ExpectContext: ""},
		testDownloadCase{CurrentRole: "manager", DownloadURL: "/downloads/b.csv", ExpectStatusCode: 404, ExpectContext: ""},
		testDownloadCase{CurrentRole: "admin", DownloadURL: "/downloads/b.csv", ExpectStatusCode: 200, ExpectContext: "Column3,Column4\n"},
		testDownloadCase{CurrentRole: "", DownloadURL: "/downloads/c.csv", ExpectStatusCode: 404, ExpectContext: ""},
		testDownloadCase{CurrentRole: "manager", DownloadURL: "/downloads/c.csv", ExpectStatusCode: 200, ExpectContext: "Column5,Column6\n"},
		testDownloadCase{CurrentRole: "admin", DownloadURL: "/downloads/c.csv", ExpectStatusCode: 200, ExpectContext: "Column5,Column6\n"},
	}

	for i, f := range filePermissions {
		if len(f.AllowRoles) > 0 {
			permission := roles.Allow(roles.Read, f.AllowRoles...)
			if err := Downloader.Get(f.FileName).SetPermission(permission); err != nil {
				t.Errorf(color.RedString(fmt.Sprintf("Download: set file permission #(%v) failure (%v)", i+1, err)))
			}
		}
	}

	for i, testCase := range testCases {
		var hasError bool
		if testCase.CurrentRole == "" {
			CurrentUser = nil
		} else {
			CurrentUser = &User{Name: "Nika", Role: testCase.CurrentRole}
		}
		req, err := http.Get(Server.URL + testCase.DownloadURL)
		if err != nil || req.StatusCode != testCase.ExpectStatusCode {
			t.Errorf(color.RedString(fmt.Sprintf("Download #(%v): status code expect %v, but get %v", i+1, testCase.ExpectStatusCode, req.StatusCode)))
			hasError = true
		}
		if testCase.ExpectContext != "" {
			body, _ := ioutil.ReadAll(req.Body)
			if string(body) != testCase.ExpectContext {
				t.Errorf(color.RedString(fmt.Sprintf("Download #(%v): context expect %v, but get %v", i+1, testCase.ExpectContext, string(body))))
				hasError = true
			}
		}
		if !hasError {
			t.Errorf(color.GreenString(fmt.Sprintf("Download #(%v): Success", i+1)))
		}
	}
}

// Helper
func clearMetaFiles(root string) {
	for _, f := range []string{"a", "b", "c"} {
		os.Remove(root + fmt.Sprintf("/test/downloads/%v.csv.meta", f))
	}
}
