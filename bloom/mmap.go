package bloom

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

type MMapFile struct {
	File *os.File
	Data []byte
	Size int
}

func OpenMMap(path string, size int) (*MMapFile, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	err = file.Truncate(int64(size))
	if err != nil {
		return nil, err
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return &MMapFile{File: file, Data: data, Size: size}, nil
}

func (m *MMapFile) Sync() error {
	return unix.Msync(m.Data, syscall.MS_SYNC)
}

func (m *MMapFile) Close() error {
	syscall.Munmap(m.Data)
	return m.File.Close()
}
