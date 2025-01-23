package fatfs

/*
#include <stdint.h>

#include "ff.h"
#include "diskio.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

const (
	C_STA_NOINIT  C.int = 0x01 /* Drive not initialized */
	C_STA_NODISK  C.int = 0x02 /* No medium in the drive */
	C_STA_PROTECT C.int = 0x04 /* Write protected */
)

var deviceMap = make(map[uint8]BlockDevice)

// RegisterBlockDevice associates a BlockDevice with a drive number.
func RegisterBlockDevice(pdrv uint8, dev BlockDevice) {
	deviceMap[pdrv] = dev
}

func UnregisterBlockDevice(pdrv uint8) {
	delete(deviceMap, pdrv)
}

//export Go_diskRead
func Go_diskRead(pdrv C.BYTE, buff *C.uchar, sector C.LBA_t, count C.uint) C.int {
	bd, ok := deviceMap[uint8(pdrv)]
	if !ok {
		return C.RES_ERROR // Some error code
	}

	// Convert the C pointer to a Go slice
	length := int(count) * 512 // Assuming sector size = 512
	buffer := (*[1 << 30]byte)(unsafe.Pointer(buff))[:length:length]

	err := bd.ReadSectors(uint64(sector), uint32(count), buffer)
	if err != nil {
		fmt.Println("diskRead error:", err)
		return C.RES_ERROR
	}
	return C.RES_OK
}

//export Go_diskWrite
func Go_diskWrite(pdrv C.BYTE, buff *C.uchar, sector C.LBA_t, count C.uint) C.int {
	bd, ok := deviceMap[uint8(pdrv)]
	if !ok {
		return C.RES_ERROR
	}

	length := int(count) * 512
	buffer := (*[1 << 30]byte)(unsafe.Pointer(buff))[:length:length]

	err := bd.WriteSectors(uint64(sector), uint32(count), buffer)
	if err != nil {
		fmt.Println("diskWrite error:", err)
		return C.RES_ERROR
	}
	return C.RES_OK
}

//export Go_diskGetSectorSize
func Go_diskGetSectorSize(pdrv C.BYTE) C.uint {
	bd, ok := deviceMap[uint8(pdrv)]
	if !ok {
		return 0
	}
	return C.uint(bd.GetSectorSize())
}

//export Go_diskGetSectorCount
func Go_diskGetSectorCount(pdrv C.BYTE) C.LBA_t {
	bd, ok := deviceMap[uint8(pdrv)]
	if !ok {
		return 0
	}
	return C.LBA_t(bd.GetSectorCount())
}

//export Go_diskInitialize
func Go_diskInitialize(pdrv C.BYTE) C.int {
	bd, ok := deviceMap[uint8(pdrv)]
	if !ok {
		return C_STA_NOINIT
	}
	if err := bd.Initialize(); err != nil {
		return C_STA_NOINIT
	}
	return 0 // success
}

//export Go_diskStatus
func Go_diskStatus(pdrv C.BYTE) C.int {
	bd, ok := deviceMap[uint8(pdrv)]
	if !ok {
		return C_STA_NOINIT
	}
	if err := bd.Status(); err != nil {
		// Could return an error status
		return C_STA_NOINIT
	}
	return 0 // OK
}
