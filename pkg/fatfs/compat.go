package fatfs

import (
	"io/fs"
	"os"

	"github.com/spf13/afero"
)

type FatAfero struct {
	*FatFs
}

var _ afero.Fs = (*FatAfero)(nil)

func (f *FatAfero) Open(name string) (afero.File, error) {
	return f.FatFs.Open(name)
}

func (f *FatAfero) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return f.FatFs.OpenFile(name, flag, perm)
}

func AsAfero(f *FatFs) *FatAfero {
	return &FatAfero{f}
}

type FatIO struct {
	*FatFs
}

var _ fs.FS = (*FatIO)(nil)

func (f *FatIO) Open(name string) (fs.File, error) {
	return f.FatFs.Open(name)
}

func AsIO(f *FatFs) *FatIO {
	return &FatIO{f}
}
