package fatfs

/*
#cgo CFLAGS: -std=gnu99

#include <stdlib.h>
#include "ff.h"
*/
import "C"
import "os"

const (
	FileResultOK                          = C.FR_OK /* (0) Succeeded */
	FileResultErr              FileResult = C.FR_DISK_ERR
	FileResultIntErr           FileResult = C.FR_INT_ERR
	FileResultNotReady         FileResult = C.FR_NOT_READY
	FileResultNoFile           FileResult = C.FR_NO_FILE
	FileResultNoPath           FileResult = C.FR_NO_PATH
	FileResultInvalidName      FileResult = C.FR_INVALID_NAME
	FileResultDenied           FileResult = C.FR_DENIED
	FileResultExist            FileResult = C.FR_EXIST
	FileResultInvalidObject    FileResult = C.FR_INVALID_OBJECT
	FileResultWriteProtected   FileResult = C.FR_WRITE_PROTECTED
	FileResultInvalidDrive     FileResult = C.FR_INVALID_DRIVE
	FileResultNotEnabled       FileResult = C.FR_NOT_ENABLED
	FileResultNoFilesystem     FileResult = C.FR_NO_FILESYSTEM
	FileResultMkfsAborted      FileResult = C.FR_MKFS_ABORTED
	FileResultTimeout          FileResult = C.FR_TIMEOUT
	FileResultLocked           FileResult = C.FR_LOCKED
	FileResultNotEnoughCore    FileResult = C.FR_NOT_ENOUGH_CORE
	FileResultTooManyOpenFiles FileResult = C.FR_TOO_MANY_OPEN_FILES
	FileResultInvalidParameter FileResult = C.FR_INVALID_PARAMETER
	FileResultReadOnly         FileResult = 99
	FileResultNotImplemented   FileResult = 0xe0 // tinyfs custom error

	TypeFAT12 Type = C.FS_FAT12
	TypeFAT16 Type = C.FS_FAT16
	TypeFAT32 Type = C.FS_FAT32
	TypeEXFAT Type = C.FS_EXFAT

	AttrReadOnly  FileAttr = C.AM_RDO
	AttrHidden    FileAttr = C.AM_HID
	AttrSystem    FileAttr = C.AM_SYS
	AttrDirectory FileAttr = C.AM_DIR
	AttrArchive   FileAttr = C.AM_ARC

	SectorSize = 512

	FileAccessRead         OpenFlag = C.FA_READ
	FileAccessWrite        OpenFlag = C.FA_WRITE
	FileAccessOpenExisting OpenFlag = C.FA_OPEN_EXISTING
	FileAccessCreateNew    OpenFlag = C.FA_CREATE_NEW
	FileAccessCreateAlways OpenFlag = C.FA_CREATE_ALWAYS
	FileAccessOpenAlways   OpenFlag = C.FA_OPEN_ALWAYS
	FileAccessOpenAppend   OpenFlag = C.FA_OPEN_APPEND
)

type OpenFlag uint

type Type uint

func (t Type) String() string {
	switch t {
	case TypeFAT12:
		return "FAT12"
	case TypeFAT16:
		return "FAT16"
	case TypeFAT32:
		return "FAT32"
	case TypeEXFAT:
		return "EXFAT"
	default:
		return "invalid/unknown"
	}
}

type FileResult uint

func (r FileResult) Error() string {
	var msg string
	switch r {
	case FileResultErr:
		msg = "(1) A hard error occurred in the low level disk I/O layer"
	case FileResultIntErr:
		msg = "(2) Assertion failed"
	case FileResultNotReady:
		msg = "(3) The physical drive cannot work"
	case FileResultNoFile:
		msg = "(4) Could not find the file"
	case FileResultNoPath:
		msg = "(5) Could not find the path"
	case FileResultInvalidName:
		msg = "(6) The path name format is invalid"
	case FileResultDenied:
		msg = "(7) Access denied due to prohibited access or directory full"
	case FileResultExist:
		msg = "(8) Access denied due to prohibited access"
	case FileResultInvalidObject:
		msg = "(9) The file/directory object is invalid"
	case FileResultWriteProtected:
		msg = "(10) The physical drive is write protected"
	case FileResultInvalidDrive:
		msg = "(11) The logical drive number is invalid"
	case FileResultNotEnabled:
		msg = "(12) The volume has no work area"
	case FileResultNoFilesystem:
		msg = "(13) There is no valid FAT volume"
	case FileResultMkfsAborted:
		msg = "(14) The f_mkfs() aborted due to any problem"
	case FileResultTimeout:
		msg = "(15) Could not get a grant to access the volume within defined period"
	case FileResultLocked:
		msg = "(16) The operation is rejected according to the file sharing policy"
	case FileResultNotEnoughCore:
		msg = "(17) LFN working buffer could not be allocated"
	case FileResultTooManyOpenFiles:
		msg = "(18) Number of open files > FF_FS_LOCK"
	case FileResultInvalidParameter:
		msg = "(19) Given parameter is invalid"
	case FileResultReadOnly:
		msg = "(99) Read-only filesystem"
	case FileResultNotImplemented:
		msg = "(e0) Feature Not Implemented"
	default:
		msg = "unknown file result error"
	}
	return "fatfs: " + msg
}

func errval(errno C.FRESULT) error {
	if errno > FileResultOK {
		return FileResult(errno)
	}
	return nil
}

type FileAttr byte

// translateFlags translates osFlags such as os.O_RDONLY into fatfs flags.
// http://elm-chan.org/fsw/ff/doc/open.html
func translateFlags(osFlags int) C.BYTE {
	var result C.BYTE
	result = C.FA_READ
	switch osFlags {
	case os.O_RDONLY:
		// r
		result = C.FA_READ
	case os.O_CREATE:
		// x
		result = C.FA_CREATE_ALWAYS
	case os.O_WRONLY:
		fallthrough
	case os.O_WRONLY | os.O_CREATE:
		fallthrough
	case os.O_WRONLY | os.O_CREATE | os.O_TRUNC:
		// w
		result = C.FA_CREATE_ALWAYS | C.FA_WRITE
	case os.O_WRONLY | os.O_CREATE | os.O_APPEND:
		// a
		result = C.FA_OPEN_APPEND | C.FA_WRITE
	case os.O_RDWR:
		// r+
		result = C.FA_READ | C.FA_WRITE
	case os.O_RDWR | os.O_CREATE | os.O_TRUNC:
		// w+
		result = C.FA_CREATE_ALWAYS | C.FA_WRITE | C.FA_READ
	case os.O_RDWR | os.O_CREATE | os.O_APPEND:
		// a+
		result = C.FA_OPEN_APPEND | C.FA_WRITE | C.FA_READ
	default:
	}
	return result
}
