package fatfs

import (
	"fmt"
	"os"
)

const (
	sectorSize = 512
)

// assert that ImageFile implements the BlockDevice interface
var _ BlockDevice = (*ImageFile)(nil)

// ImageFile is a struct that implements the BlockDevice interface.
type ImageFile struct {
	file *os.File
}

// NewImageFile initializes an ImageFile by opening or creating a file at path.
// Adjust flags and permissions as needed.
func NewImageFile(path string) (*ImageFile, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &ImageFile{file: f}, nil
}

// Initialize might be a no-op for a simple file, or you could place
// additional logic here if needed.
func (img *ImageFile) Initialize() error {
	// For example, you might want to check file size or do other init tasks
	return nil
}

// Status might also be a no-op, or you could do checks on the file's state.
func (img *ImageFile) Status() error {
	// Example: verify the file handle is still valid
	if img.file == nil {
		return fmt.Errorf("file is not open")
	}
	return nil
}

// ReadSectors reads `count` sectors from the file at the sector index `sector`
// into the buffer `buff`.
func (img *ImageFile) ReadSectors(sector uint64, count uint32, buff []byte) error {
	if img.file == nil {
		return fmt.Errorf("file is not open")
	}

	// Calculate byte offset in the file
	offset := int64(sector * sectorSize)
	length := int64(count * sectorSize)

	// Ensure the buffer is large enough
	if int64(len(buff)) < length {
		return fmt.Errorf("buffer too small: need %d bytes, got %d", length, len(buff))
	}

	// Seek to the correct offset
	_, err := img.file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	// Read data into buff
	n, err := img.file.Read(buff[:length])
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}
	if int64(n) != length {
		return fmt.Errorf("short read: expected %d bytes, got %d", length, n)
	}

	return nil
}

// WriteSectors writes `count` sectors from the buffer `buff` to the file
// at the sector index `sector`.
func (img *ImageFile) WriteSectors(sector uint64, count uint32, buff []byte) error {
	if img.file == nil {
		return fmt.Errorf("file is not open")
	}

	// Calculate byte offset in the file
	offset := int64(sector * sectorSize)
	length := int64(count * sectorSize)

	// Ensure the buffer has enough data
	if int64(len(buff)) < length {
		return fmt.Errorf("buffer too small: need %d bytes, got %d", length, len(buff))
	}

	// Seek to the correct offset
	_, err := img.file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	// Write data from buff
	n, err := img.file.Write(buff[:length])
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}
	if int64(n) != length {
		return fmt.Errorf("short write: expected %d bytes, wrote %d", length, n)
	}

	return nil
}

// GetSectorSize returns the sector size of the file.
func (img *ImageFile) GetSectorSize() uint64 {
	return sectorSize
}

// GetSectorCount returns the number of sectors in the file.
func (img *ImageFile) GetSectorCount() uint64 {
	if img.file == nil {
		return 0
	}
	info, err := img.file.Stat()
	if err != nil {
		return 0
	}
	return uint64(info.Size() / sectorSize)
}

// Close should be called when you're done with the ImageFile
func (img *ImageFile) Close() error {
	if img.file == nil {
		return nil
	}
	err := img.file.Close()
	img.file = nil
	return err
}
