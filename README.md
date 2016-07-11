# Filebox

Filebox is component that used to make a local file able to be download and provide access control.

You could choose filebox to satisfy below scenarios

Scenario 1:

* You would like to make file `~/documents/files/users.csv` able to be download via link `http://127.0.0.1/downloads/users.csv`
* Only Admin user able to download this file

Scenario 2:

* You create a folder at `~/exchanges`
* Only Admin user could access files in this folder

[![GoDoc](https://godoc.org/github.com/qor/filebox?status.svg)](https://godoc.org/github.com/qor/filebox)

## Usage

```go
import (
	"github.com/qor/filebox"
	"github.com/qor/roles"
	"net/http"
	"string"
)

func main() {
	mux := http.NewServeMux()
	Filebox = filebox.New("/home/qor/project/downloads")
	Filebox.MountTo("/downloads", mux)

	// Assert folder downloads has file users.csv
	// then you could download this file by http://127.0.0.1:7000/downloads/users.csv

	// Add permission for users.csv, limit to only admin user able to access
    permission := roles.Allow(roles.Read, "admin")
    newFile := Filebox.AccessFile("users.csv")
    newFile.SetPermission(permission)

    // Add permission for a specific directory
    dir := Dir.AccessDir("/exchanges")
    dir.SetPermission(permission)
    // Create a new file and this file will have same permission setting as directory
    dir.WriteFile("products.csv", strings.NewReader("Content"))
}

```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
