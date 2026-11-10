package testfixtures

import (
	"embed"
	"io/fs"
)

//go:embed testdata/*.sql
var files embed.FS

var FS = migrations()

func migrations() fs.FS {
	src, err := fs.Sub(files, "testdata")
	if err != nil {
		panic(err)
	}
	return src
}
