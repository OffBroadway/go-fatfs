package fatfs

/*
#cgo CFLAGS: -std=gnu99

#include <stdlib.h>
#include "ff.h"

static int GET_MACRO_FF_VOLUMES() { return FF_VOLUMES; }

// A helper function so we can create a FATFS struct in C and return a pointer
FATFS* allocate_fatfs() {
    FATFS* fs = (FATFS*)malloc(sizeof(FATFS));
    return fs;
}

FIL* allocate_fil(void) {
    FIL* fil = (FIL*)malloc(sizeof(FIL));
	return fil;
}

FF_DIR* allocate_dir(void) {
    FF_DIR* dir = (FF_DIR*)malloc(sizeof(FF_DIR));
	return dir;
}

FRESULT unmount_fs(const TCHAR* path) {
	return f_unmount(path);
}

FSIZE_t fatfs_tell(FIL* fp) {
    return f_tell(fp);
}

*/
import "C"
import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/spf13/afero"
)

// FatFs holds a pointer to the FATFS structure allocated in C.
type FatFs struct {
	fs *C.FATFS

	volNumber uint8
	volPrefix string
	openFiles map[string]*FatFile
}

// Fil is a Go wrapper around the FIL struct from FatFs
type FatFile struct {
	fs  *FatFs
	fil *C.FIL
	dir *C.FF_DIR

	info FileInfo

	writeAppendMode bool
}

type FileInfo struct {
	name    string
	size    int64
	isDir   bool
	modTime time.Time
	mode    os.FileMode
	sys     interface{}
}

func (fi FileInfo) Name() string       { return fi.name }
func (fi FileInfo) Size() int64        { return fi.size }
func (fi FileInfo) IsDir() bool        { return fi.isDir }
func (fi FileInfo) ModTime() time.Time { return fi.modTime }
func (fi FileInfo) Mode() os.FileMode  { return fi.mode }
func (fi FileInfo) Sys() interface{}   { return fi.sys }

var _ os.FileInfo = FileInfo{}

// NewFatFs allocates a new FATFS struct in C.
func NewFatFs(volume int) (*FatFs, error) {
	if volume >= int(C.GET_MACRO_FF_VOLUMES()) {
		return nil, fmt.Errorf("volume number exceeds maximum: %d", int(C.GET_MACRO_FF_VOLUMES()))
	}

	fs := C.allocate_fatfs()
	if fs == nil {
		return nil, fmt.Errorf("failed to allocate FATFS")
	}
	obj := &FatFs{
		fs:        fs,
		volNumber: uint8(volume),
		volPrefix: fmt.Sprintf("%d:", volume),
		openFiles: make(map[string]*FatFile),
	}
	return obj, nil
}

func (f *FatFs) Name() string {
	fmt.Println("CALL Name")
	return "FatFs"
}

// Mount calls f_mount internally.
func (f *FatFs) Mount(blk BlockDevice) error {
	cpath := C.CString(f.volPrefix)
	defer C.free(unsafe.Pointer(cpath))

	RegisterBlockDevice(f.volNumber, blk)

	res := C.f_mount(f.fs, (*C.TCHAR)(unsafe.Pointer(cpath)), C.BYTE(0))
	if res != 0 {
		return fmt.Errorf("f_mount error code: %d", res)
	}

	// TODO: is this redundant?
	f.openFiles = make(map[string]*FatFile)

	return nil
}

func (f *FatFs) Unmount() error {
	for _, file := range f.openFiles {
		file.Close() // TODO: handle errors
	}

	cpath := C.CString(f.volPrefix)
	defer C.free(unsafe.Pointer(cpath))

	res := C.unmount_fs((*C.TCHAR)(unsafe.Pointer(cpath)))
	if res != 0 {
		return fmt.Errorf("f_unmount error code: %d", res)
	}

	UnregisterBlockDevice(f.volNumber)
	return nil
}

func (f *FatFs) Chmod(name string, mode os.FileMode) error {
	fmt.Println("STUB Chmod", name, mode)
	// return os.ErrPermission
	return nil
}

func (f *FatFs) Chown(name string, uid, gid int) error {
	fmt.Println("STUB Chown", name, uid, gid)
	// return os.ErrPermission
	return nil
}

func (f *FatFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	fmt.Println("STUB Chtimes", name, atime, mtime)
	return nil
}

func (f *FatFs) Open(path string) (*FatFile, error) {
	fmt.Println("CALL Open", path)
	return f.OpenFile(path, os.O_RDONLY, 0o644)
}

func (f *FatFs) OpenFile(path string, flags int, perm os.FileMode) (*FatFile, error) {
	fmt.Println("CALL OpenFile", path, flags, uint32(perm))
	file := &FatFile{fs: f}
	file.writeAppendMode = isWriteMode(flags) && isAppendMode(flags)

	cpath := C.CString(f.volPrefix + path)
	defer C.free(unsafe.Pointer(cpath))

	isDir := false
	infos, err := f.Stat(path)
	if err == nil {
		isDir = infos.IsDir()
		file.info = *infos.(*FileInfo)
	}

	var errno C.FRESULT
	if path == "/" || isDir {
		fmt.Println("Opening directory:", path)
		file.dir = C.allocate_dir()
		if file.dir == nil {
			return nil, fmt.Errorf("failed to allocate DIR")
		}
		errno = C.f_opendir(file.dir, (*C.TCHAR)(unsafe.Pointer(cpath)))
	} else {
		fmt.Println("Opening file:", path)
		file.fil = C.allocate_fil()
		if file.fil == nil {
			fmt.Println("Failed to allocate FIL")
			return nil, fmt.Errorf("failed to allocate FIL")
		}
		errno = C.f_open(file.fil, (*C.TCHAR)(unsafe.Pointer(cpath)), translateFlags(flags))
	}

	// check to make sure f_open/f_opendir didn't produce an error
	if err := errval(errno); err != nil {
		fmt.Println("f_open/f_opendir error:", err)
		if file.dir != nil {
			C.free(unsafe.Pointer(file.dir))
			file.dir = nil
		}
		if file.fil != nil {
			C.free(unsafe.Pointer(file.fil))
			file.fil = nil
		}
		return nil, err
	}

	if file.info.name == "" {
		fmt.Println("File info not found, getting from path")

		// fill in the file info
		infos, err = f.Stat(path)
		if err != nil {
			fmt.Println("OpenFile Stat error:", err)
			return nil, err
		}
		file.info = *infos.(*FileInfo)
	}

	// file handle was initialized successfully
	f.openFiles[path] = file
	return file, nil

}

func (f *FatFs) Create(name string) (afero.File, error) {
	fmt.Println("CALL Create", name)
	return f.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
}

func (f *FatFs) Remove(name string) error {
	fmt.Println("CALL Remove", name)

	cpath := C.CString(f.volPrefix + name)
	defer C.free(unsafe.Pointer(cpath))

	return errval(C.f_unlink(cpath))
}

func (f *FatFs) RemoveAll(path string) error {
	if info, err := f.Stat(path); err == nil && !info.IsDir() {
		return f.Remove(path)
	}

	fmt.Println("STUB RemoveAll (dir)", path)
	return os.ErrInvalid
}

func (f *FatFs) Rename(oldname, newname string) error {
	fmt.Println("STUB Rename")
	return nil
}

func (f *FatFs) Mkdir(name string, perm os.FileMode) error {
	fmt.Println("STUB Mkdir", name, perm)
	return nil
}

func (f *FatFs) MkdirAll(path string, perm os.FileMode) error {
	fmt.Println("CALL MkdirAll", path, perm)

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Split the path into components
	currentPath := ""
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			continue
		}

		// Build the current path incrementally
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath += "/" + part
		}

		// Check if the directory exists
		_, err := f.Stat(currentPath)
		if err != nil {
			// If the error is not "does not exist", return the error
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to check directory %s: %w", currentPath, err)
			}

			// Directory does not exist; attempt to create it
			fmt.Println("Creating directory:", currentPath)
			err = f.Mkdir(currentPath, perm)
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %w", currentPath, err)
			}
		}
	}

	return nil
}

func (f *FatFs) Stat(path string) (os.FileInfo, error) {
	fmt.Printf("CALL Stat [%s]\n", path)
	if path == "/" || path == "." || path == "" {
		info := FileInfo{
			name:    "/",
			size:    int64(512 * 1024 * 1024),
			isDir:   true,
			modTime: time.Unix(0, 0),
			mode:    os.ModeDir | os.ModePerm | os.ModeDevice,
		}
		return &info, nil
	}

	cpath := C.CString(f.volPrefix + path)
	defer C.free(unsafe.Pointer(cpath))

	info := C.FILINFO{}
	if err := errval(C.f_stat(cpath, &info)); err != nil {
		fmt.Println("f_stat error:", err)
		if errors.Is(err, FileResultNoFile) {
			return nil, os.ErrNotExist
		} else if errors.Is(err, FileResultInvalidObject) {
			return nil, os.ErrInvalid
		}
		return nil, err
	}

	fname := C.GoString(&info.fname[0])
	infos := FileInfo{
		name:    fname,
		size:    int64(info.fsize),
		isDir:   info.fattrib&C.AM_DIR != 0,
		modTime: time.Unix(0, 0),
		mode:    os.ModePerm,
	}

	return &infos, nil
}

// File methods

// Name returns the name of the file as presented to OpenFile
func (f *FatFile) Name() string {
	return f.info.name
}

func (f *FatFile) readDir() (infos []os.FileInfo, err error) {
	if !f.info.IsDir() {
		return nil, FileResultInvalidObject
	}
	for {
		info := C.FILINFO{}
		if err := errval(C.f_readdir(f.dir, &info)); err != nil {
			return nil, err
		}
		fname := C.GoString(&info.fname[0])
		if fname == "" {
			return infos, nil
		}

		infos = append(infos, &FileInfo{
			name:    fname,
			size:    int64(info.fsize),
			isDir:   info.fattrib&C.AM_DIR != 0,
			modTime: time.Unix(0, 0),
			mode:    os.ModePerm,
		})
	}
}

func (f *FatFile) Readdir(count int) ([]os.FileInfo, error) {
	fmt.Println("CALL Readdir", count)
	res, err := f.readDir()
	if err != nil {
		return nil, err
	}
	if count > 0 {
		if len(res) > count {
			res = res[:count]
		}
	}
	return res, nil
}

func (f *FatFile) Readdirnames(n int) (names []string, err error) {
	infos, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		names = append(names, info.Name())
	}
	return names, nil
}

// Read from a file
func (f *FatFile) Read(data []byte) (int, error) {
	// fmt.Println("CALL Read", len(data))
	if f.info.IsDir() {
		return 0, FileResultInvalidObject
	}
	var br, btw C.UINT = 0, C.UINT(len(data))
	res := C.f_read(f.fil, unsafe.Pointer(&data[0]), C.UINT(len(data)), &br)
	if res != 0 {
		fmt.Println("f_read error code:", errval(res))
		return 0, fmt.Errorf("f_read error code: %d", res)
	}
	fmt.Println("f_read bytes read:", br)
	if br == 0 && btw > 0 {
		return 0, io.EOF
	}
	return int(br), nil
}

// Write to a file
func (f *FatFile) Write(buf []byte) (int, error) {
	// fmt.Println("CALL Write", len(buf))
	if f.info.IsDir() {
		return 0, FileResultInvalidObject
	}

	bufptr := unsafe.Pointer(&buf[0])
	var bw, btw C.UINT = 0, C.UINT(len(buf))
	errno := C.f_write(f.fil, bufptr, btw, &bw)
	if err := errval(errno); err != nil {
		return int(bw), err
	}

	if bw < btw {
		fmt.Printf("DEBUG: Volume Full %d < %d\n", bw, btw)
		return int(bw), errors.New("volume is full")
	}

	return int(bw), nil
}

func (f *FatFile) WriteAt(buf []byte, offset int64) (n int, err error) {
	// fmt.Println("CALL WriteAt", len(buf), offset)
	if f.info.IsDir() {
		return 0, FileResultInvalidObject
	}

	oldPos := C.fatfs_tell(f.fil)
	defer C.f_lseek(f.fil, oldPos)

	bufptr := unsafe.Pointer(&buf[0])
	var bw, btw C.UINT = 0, C.UINT(len(buf))
	errno := C.f_lseek(f.fil, C.FSIZE_t(offset))
	if err := errval(errno); err != nil {
		return int(bw), err
	}

	errno = C.f_write(f.fil, bufptr, btw, &bw)
	if err := errval(errno); err != nil {
		return int(bw), err
	}

	if bw < btw {
		return int(bw), errors.New("volume is full")
	}
	return int(bw), nil
}

func (f *FatFile) WriteString(s string) (n int, err error) {
	return f.Write([]byte(s))
}

// Seek changes the position of the file
func (f *FatFile) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case io.SeekStart:
		// pass
		fmt.Println("SEEK_START", offset)
	case io.SeekCurrent:
		offset += int64(C.fatfs_tell(f.fil))
		fmt.Println("SEEK_CURRENT", offset)
	case io.SeekEnd:
		if f.writeAppendMode {
			offset += int64(C.fatfs_tell(f.fil))
			fmt.Println("SEEK_END_APPEND", offset)
		} else {
			offset += f.info.size
			fmt.Println("SEEK_END", offset)
		}
	default:
		return -1, FileResultInvalidParameter
	}
	errno := C.f_lseek(f.fil, C.FSIZE_t(offset))
	if err := errval(errno); err != nil {
		return -1, err
	}
	return offset, nil
}

func (f *FatFile) ReadAt(buf []byte, offset int64) (n int, err error) {
	if f.info.IsDir() {
		return 0, FileResultInvalidObject
	}
	bufptr := unsafe.Pointer(&buf[0])
	var br, btr C.UINT = 0, C.UINT(len(buf))
	errno := C.f_lseek(f.fil, C.FSIZE_t(offset))
	if err := errval(errno); err != nil {
		return int(br), err
	}
	errno = C.f_read(f.fil, bufptr, btr, &br)
	if err := errval(errno); err != nil {
		return int(br), err
	}
	if br == 0 && btr > 0 {
		return 0, io.EOF
	}
	return int(br), nil
}

func (f *FatFile) Stat() (os.FileInfo, error) {
	return f.info, nil
}

// Sync the file
func (f *FatFile) Sync() error {
	return errval(C.f_sync(f.fil))
}

// Truncates the size of the file to the specified size
//
// Returns a negative error code on failure.
func (f *FatFile) Truncate(size int64) error {
	// seek then f_truncate
	errno := C.f_lseek(f.fil, C.FSIZE_t(size))
	if err := errval(errno); err != nil {
		return err
	}
	errno = C.f_truncate(f.fil)
	if err := errval(errno); err != nil {
		return err
	}

	return nil
}

// Close the file
func (f *FatFile) Close() error {
	fmt.Println("CALL Close", f.info.name)

	delete(f.fs.openFiles, f.info.name)

	var errno C.FRESULT
	if f.fil != nil || f.dir != nil {
		if f.info.IsDir() {
			errno = C.f_closedir(f.dir)
		} else {
			errno = C.f_close(f.fil)
		}
	}
	return errval(errno)
}
