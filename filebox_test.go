package filebox_test

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/qor/admin"
	"github.com/qor/filebox"
	"github.com/qor/qor"
	"github.com/qor/roles"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var Filebox *filebox.Filebox
var Admin *admin.Admin
var Server *httptest.Server
var CurrentUser *User
var Root string

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
	Root, _ = os.Getwd()
	mux := http.NewServeMux()
	Server = httptest.NewServer(mux)
	CurrentUser = &User{Name: "user", Role: "normal_user"}
	roles.Register("admin", func(req *http.Request, currentUser interface{}) bool {
		return currentUser.(*User) != nil && currentUser.(*User).Role == "admin"
	})
	roles.Register("manager", func(req *http.Request, currentUser interface{}) bool {
		return currentUser.(*User) != nil && currentUser.(*User).Role == "manager"
	})

	Filebox = filebox.New(Root + "/test/filebox")
	Filebox.MountTo(mux)
	Filebox.SetAuth(AdminAuth{})
}

func reset() {
	clearFiles()
}

// Test download cases
type filePermission struct {
	DirPermssion *roles.Permission
	FileName     string
	AllowRoles   []string
}

type testDownloadCase struct {
	CurrentRole      string
	DownloadURL      string
	ExpectStatusCode int
	ExpectContext    string
}

func TestDownloads(t *testing.T) {
	reset()
	filePermissions := []filePermission{
		filePermission{FileName: "a.csv", AllowRoles: []string{}},
		filePermission{FileName: "b.csv", AllowRoles: []string{"admin"}},
		filePermission{FileName: "c.csv", AllowRoles: []string{"manager", "admin"}},
		filePermission{FileName: "translations/en.csv", AllowRoles: []string{"manager", "admin"}},
		// File doesn't set permission, but Dir set
		filePermission{
			DirPermssion: roles.Allow(roles.Read, "admin"),
			FileName:     "translations/users.csv",
			AllowRoles:   []string{},
		},
		// File set permission and Dir set permission too, File's permission will override Dir's permission
		filePermission{
			DirPermssion: roles.Allow(roles.Read, "admin"),
			FileName:     "translations/products.csv",
			AllowRoles:   []string{"manager", "admin"},
		},
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
		testDownloadCase{CurrentRole: "", DownloadURL: "/downloads/translations/en.csv", ExpectStatusCode: 404, ExpectContext: ""},
		testDownloadCase{CurrentRole: "manager", DownloadURL: "/downloads/translations/en.csv", ExpectStatusCode: 200, ExpectContext: "Key,Value\n"},
		testDownloadCase{CurrentRole: "admin", DownloadURL: "/downloads/translations/en.csv", ExpectStatusCode: 200, ExpectContext: "Key,Value\n"},
		testDownloadCase{CurrentRole: "", DownloadURL: "/downloads/translations/users.csv", ExpectStatusCode: 404, ExpectContext: ""},
		testDownloadCase{CurrentRole: "manager", DownloadURL: "/downloads/translations/users.csv", ExpectStatusCode: 404, ExpectContext: ""},
		testDownloadCase{CurrentRole: "admin", DownloadURL: "/downloads/translations/users.csv", ExpectStatusCode: 200, ExpectContext: "ID,Name\n"},
		testDownloadCase{CurrentRole: "", DownloadURL: "/downloads/translations/products.csv", ExpectStatusCode: 404, ExpectContext: ""},
		testDownloadCase{CurrentRole: "manager", DownloadURL: "/downloads/translations/products.csv", ExpectStatusCode: 200, ExpectContext: "ID,Code\n"},
		testDownloadCase{CurrentRole: "admin", DownloadURL: "/downloads/translations/products.csv", ExpectStatusCode: 200, ExpectContext: "ID,Code\n"},
	}

	for i, f := range filePermissions {
		if len(f.AllowRoles) > 0 {
			permission := roles.Allow(roles.Read, f.AllowRoles...)
			newFile := Filebox.AccessFile(f.FileName)
			if err := newFile.SetPermission(permission); err != nil {
				t.Errorf(color.RedString(fmt.Sprintf("Filebox: set file permission #(%v) failure (%v)", i+1, err)))
			}
		}
		if f.DirPermssion != nil {
			newFile := Filebox.AccessFile(f.FileName)
			newFile.Dir.SetPermission(f.DirPermssion)
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
			fmt.Printf(color.GreenString("Download #%v: Success\n", i+1))
		}
	}
}

// Test Put file
type testPutFileCase struct {
	FilePath       string
	Context        string
	ExpectSavePath string
	ExpectContext  string
}

func TestPutFile(t *testing.T) {
	reset()
	testCases := []testPutFileCase{
		testPutFileCase{
			FilePath:       "new1.csv",
			Context:        "String: Hello world!",
			ExpectSavePath: "/test/filebox/new1.csv",
			ExpectContext:  "Hello world!",
		},
		testPutFileCase{
			FilePath:       "new2.csv",
			Context:        "File: a.csv",
			ExpectSavePath: "/test/filebox/new2.csv",
			ExpectContext:  "Column1,Column2\n",
		},
		testPutFileCase{
			FilePath:       "jobs/translation.csv",
			Context:        "File: a.csv",
			ExpectSavePath: "/test/filebox/jobs/translation.csv",
			ExpectContext:  "Column1,Column2\n",
		},
		testPutFileCase{
			FilePath:       "jobs/translations/1/file.csv",
			Context:        "File: a.csv",
			ExpectSavePath: "/test/filebox/jobs/translations/1/file.csv",
			ExpectContext:  "Column1,Column2\n",
		},
	}
	for i, testCase := range testCases {
		var reader io.Reader
		if strings.HasPrefix(testCase.Context, "String:") {
			reader = strings.NewReader(strings.Replace(testCase.Context, "String: ", "", 1))
		} else {
			fileName := strings.Replace(testCase.Context, "File: ", "", 1)
			reader, _ = os.Open(Root + "/test/filebox/" + fileName)
		}
		newFile := Filebox.AccessFile(testCase.FilePath)
		err := newFile.Write(reader)
		if err != nil {
			t.Errorf(color.RedString(fmt.Sprintf("Put file #%v: create file %v failure, get error %v", i+1, testCase.ExpectSavePath, err)))
		}
		permission := roles.Allow(roles.Read, "admin")
		err = newFile.SetPermission(permission)
		if err != nil {
			t.Errorf(color.RedString(fmt.Sprintf("Put file #%v: set permission to file %v failure, get error %v", i+1, testCase.ExpectSavePath, err)))
		}
		var hasError bool
		if _, err := os.Stat(Root + testCase.ExpectSavePath); os.IsNotExist(err) {
			hasError = true
			t.Errorf(color.RedString(fmt.Sprintf("Put file #%v: should create %v", i+1, testCase.ExpectSavePath)))
		} else {
			context, _ := ioutil.ReadFile(Root + testCase.ExpectSavePath)
			if string(context) != testCase.ExpectContext {
				t.Errorf(color.RedString(fmt.Sprintf("Put file #%v: context should be as %v, but get %v", i+1, testCase.ExpectContext, string(context))))
			}
		}
		if _, err := os.Stat(Root + testCase.ExpectSavePath + ".meta"); os.IsNotExist(err) {
			hasError = true
			t.Errorf(color.RedString(fmt.Sprintf("Put file #%v: should create %v.meta", i+1, testCase.ExpectSavePath)))
		}
		if !hasError {
			fmt.Printf(color.GreenString("Put file #%v: Success\n", i+1))
		}
	}
	clearFiles()
}

// Test Set permission to a folder and write file
func TestDirPutFile(t *testing.T) {
	reset()
	dir := Filebox.AccessDir("private")
	permission := roles.Allow(roles.Read, "admin")
	dir.SetPermission(permission)
	dir.WriteFile("a.csv", strings.NewReader("Hello"))
	CurrentUser = &User{Name: "Nika", Role: ""}
	req, err := http.Get(Server.URL + "/downloads/private/a.csv")
	if err != nil || req.StatusCode != 404 {
		t.Errorf(color.RedString(fmt.Sprintf("Dir: status code expect 404, but get %v", req.StatusCode)))
	}
	CurrentUser = &User{Name: "Nika", Role: "admin"}
	req, err = http.Get(Server.URL + "/downloads/private/a.csv")
	if err != nil || req.StatusCode != 200 {
		t.Errorf(color.RedString(fmt.Sprintf("Dir: status code expect 200, but get %v", req.StatusCode)))
	}
	clearFiles()
}

// Helper
func clearFiles() {
	for _, f := range []string{"a", "b", "c", "new1", "new2"} {
		os.Remove(Root + fmt.Sprintf("/test/filebox/%v.csv.meta", f))
	}
	os.Remove(Root + "/test/filebox/new1.csv")
	os.Remove(Root + "/test/filebox/new2.csv")
	os.Remove(Root + "/test/filebox/translations/en.csv.meta")
	os.Remove(Root + "/test/filebox/translations/products.csv.meta")
	os.Remove(Root + "/test/filebox/translations/.meta")
	os.RemoveAll(Root + "/test/filebox/jobs")
	os.RemoveAll(Root + "/test/filebox/private")
}
