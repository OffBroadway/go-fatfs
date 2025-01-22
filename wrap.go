package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

type FS struct {
	afero.Fs
}

func newFS(fs afero.Fs) *FS {
	return &FS{
		Fs: fs,
	}
}

func (f *FS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	fmt.Println("webdav Mkdir")
	return f.Fs.Mkdir(name, perm)
}

func (f *FS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	fmt.Println("webdav OpenFile")
	return f.Fs.OpenFile(name, flag, perm)
}

func (f *FS) RemoveAll(ctx context.Context, name string) error {
	fmt.Println("webdav RemoveAll")
	return f.Fs.RemoveAll(name)
}

func (f *FS) Rename(ctx context.Context, oldName, newName string) error {
	fmt.Println("webdav Rename")
	return f.Fs.Rename(oldName, newName)
}

func (f *FS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fmt.Println("webdav Stat")
	fileInfo, err := f.Fs.Stat(name)
	if err != nil {
		return nil, err
	}
	return fileInfo, err
}

func newHandler(fs webdav.FileSystem, prefix string) http.Handler {
	return &webdav.Handler{
		Prefix:     prefix,
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
	}
}

func Serve(listener net.Listener, fs afero.Fs) error {
	// memfs := webdav.NewMemFS()
	// h := newHandler(memfs, "/mount")
	h := newHandler(newFS(fs), "/mount")
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	server := &http.Server{
		Handler:  handlers.LoggingHandler(os.Stdout, h),
		ErrorLog: logger,
	}
	return server.Serve(listener)
}
