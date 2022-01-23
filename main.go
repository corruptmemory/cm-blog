package main

import (
	"embed"
	_ "embed"
	"io/fs"
	"log"
	"net/http"
	"path"
)

// content holds our static web server content.
//go:embed content
var content embed.FS

type ReRootedFS struct {
	prefix string
	inner  embed.FS
}

func (r *ReRootedFS) Open(name string) (fs.File, error) {
	return r.inner.Open(path.Join(r.prefix, name))
}

func (r *ReRootedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return r.inner.ReadDir(path.Join(r.prefix, name))
}

// ReadFile reads and returns the content of the named file.
func (r *ReRootedFS) ReadFile(name string) ([]byte, error) {
	return r.inner.ReadFile(path.Join(r.prefix, name))
}

func main() {
	log.Println("Hello from the Corruptmemory Blog server")
	newHome := &ReRootedFS{
		prefix: "content",
		inner:  content,
	}
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.FS(newHome))))
}
