# Filebox

Filebox provides access permission control for files and directories.

You could utilize Filebox to satisfy the following scenarios:

Scenario 1:

* You would like to make a file, let's say `~/documents/files/users.csv`, downloadable via  the URL `http://127.0.0.1/downloads/users.csv`.
* restrict downloads of that file to Admin users.

Scenario 2:

* Restricting access to files within a folder, let's say `~/exchanges`, to Admin users.

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

  // Mount filebox into `/downloads`
  Filebox := filebox.New("/home/qor/project/downloads")
  Filebox.MountTo("/downloads", mux)

  // Assert folder downloads has file users.csv
  // then you could download this file by http://127.0.0.1:7000/downloads/users.csv

  // Add permission for users.csv, limit to only admin user able to access
  permission := roles.Allow(roles.Read, "admin")
  userFile := Filebox.AccessFile("users.csv")
  userFile.SetPermission(permission)
  // read content from file `users.csv`
  fileContentReader, err := userFile.Read()
  // write content for file `users.csv`
  userFile.Write(fileContentReader)

  // Add permission for a specific directory
  exchangesDir := Dir.AccessDir("/exchanges")
  exchangesDir.SetPermission(permission)
  // Create a new file and it will use directory's permission if it hasn't define its own
  exchangesDir.WriteFile("products.csv", strings.NewReader("Content"))
}
```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
