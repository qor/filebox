# Downloader

Downloader is component that used to make a local file able to be download and provide access control.

Scenario:

* You would like to make file `~/documents/files/users.csv` able to be download via link `http://127.0.0.1/downloads/users.csv`
* Only Admin user able to download this file

Then you can use downloader to satisfy above scenario

[![GoDoc](https://godoc.org/github.com/qor/downloader?status.svg)](https://godoc.org/github.com/qor/downloader)

## Usage

```go
import (
	"github.com/qor/downloader"
	"github.com/qor/roles"
	"net/http"
	"string"
)

func main() {
	mux := http.NewServeMux()
	Downloader = downloader.New("/home/qor/project/downloads")
	Downloader.MountTo(mux)

	// Assert folder downloads has file users.csv
	// then you could download this file by http://127.0.0.1:7000/downloads/users.csv

	// Add permission for users.csv, limit to only admin user able to access
	permission := roles.Allow(roles.Read, "admin")
	Downloader.Get("users.csv").SetPermission(permission)

	// Save a io.Reader's context to a file and add permission
	reader := strings.NewReader("blabla")
	newDownloader, _ := Downloader.Put("new_file.csv", reader)
	newDownloader.SetPermission(permission)
}

```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
