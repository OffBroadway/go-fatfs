package fatfs

// BlockDevice is the interface we want to match.
type BlockDevice interface {
	ReadSectors(sector uint64, count uint32, buff []byte) error
	WriteSectors(sector uint64, count uint32, buff []byte) error
	GetSectorSize() uint64
	GetSectorCount() uint64
	Initialize() error
	Status() error
}
