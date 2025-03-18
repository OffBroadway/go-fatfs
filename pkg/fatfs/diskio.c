#include <time.h>

#include "ff.h"       // For DRESULT, DSTATUS, etc.
#include "diskio.h"   // If you have a separate diskio.h
#include "_cgo_export.h" // Magic cgo-generated header to call Go functions

/*
 * FatFs will call this function to read sectors from the storage device.
 * We'll forward it to a Go function: Go_diskRead()
 */
DRESULT disk_read (
    BYTE pdrv,     /* Physical drive number */
    BYTE *buff,    /* Data buffer to store read data */
    LBA_t sector,  /* Sector address (LBA) */
    UINT count     /* Number of sectors to read */
)
{
    return (DRESULT)Go_diskRead(pdrv, buff, sector, count);
}

/*
 * FatFs calls this to write sectors to the storage device.
 */
DRESULT disk_write (
    BYTE pdrv,
    const BYTE *buff,
    LBA_t sector,
    UINT count
)
{
    return (DRESULT)Go_diskWrite(pdrv, (BYTE*)buff, sector, count);
}

/*
 * FatFs calls this for various initialization/status checks.
 */
DSTATUS disk_initialize (BYTE pdrv)
{
    return (DSTATUS)Go_diskInitialize(pdrv);
}

DSTATUS disk_status (BYTE pdrv)
{
    return (DSTATUS)Go_diskStatus(pdrv);
}


DRESULT disk_ioctl (
	BYTE pdrv,		/* Physical drive nmuber (0..) */
	BYTE cmd,		/* Control code */
	void *buff		/* Buffer to send/receive control data */
)
{
    switch(cmd) {
        case CTRL_SYNC:
            return RES_OK;
            break;
            
        case GET_SECTOR_COUNT:
            if(!buff) return RES_PARERR;
            *(LBA_t*)buff = Go_diskGetSectorCount(pdrv);
            break;
            
        case GET_SECTOR_SIZE:
            if(!buff) return RES_PARERR;
#if FF_MAX_SS != FF_MIN_SS
			*((WORD*)buff) = Go_diskGetSectorSize(pdrv);
#else
			*((WORD*)buff) = FF_MIN_SS;
#endif
            break;
            
        case GET_BLOCK_SIZE:
            return RES_PARERR;
            break;
            
        case CTRL_TRIM:
            return RES_PARERR;
            break;
            
        default:
            return RES_PARERR;
    }

    return RES_OK;
}

#if defined(_WIN32) || defined(_WIN64)
#define localtime_r(a,b) (localtime_s(b,a) ? 0 : b)
#endif

DWORD get_fattime(void)
{ 
	struct tm tm;
	time_t now = time(NULL);
	if (localtime_r(&now, &tm) != NULL) {
		DWORD fattime =
			/* bit31:25: Year origin from the 1980 (0..127, e.g. 37 for 2017) */
			(((tm.tm_year - 80) & 0x7f) << 25) |
			/* bit24:21: Month (1..12) */
			(((tm.tm_mon + 1) & 0xf) << 21) |
			/* bit20:16: Day of the month (1..31) */
			((tm.tm_mday & 0x1f) << 16) |
			/* bit15:11: Hour (0..23)) */
			((tm.tm_hour & 0x1f) << 11) |
			/* bit10:5: Minute (0..59) */
			((tm.tm_min & 0x3f) << 5) |
			/* bit4:0 Second / 2 (0..29, e.g. 25 for 50) */
			((tm.tm_sec & 0x3f) / 2);
		//printf("get_fattime %x\n", fattime);
		return fattime;
	} else
		return 1; 
}