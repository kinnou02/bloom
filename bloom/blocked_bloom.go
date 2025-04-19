package bloom

import (
	"encoding/binary"
	"errors"
	"math"
	"os"

	"golang.org/x/sys/unix"
)

const (
	blockSize   = 64     // 64 bytes = 512 bits
	headerSize  = 16     // magic(4) + version(4) + k(4) + blocks(4)
	magicBytes  = "BBF1" // 4 bytes identifier
	fileVersion = 1
)

type BlockedBloomFilter struct {
	file    *os.File
	data    []byte
	k       uint32
	nBlocks uint32
	hasher  Hasher
}

func CreateBlockedBloomFilter(path string, nElements int, targetFP float64) (*BlockedBloomFilter, error) {
	bitsNeeded := int(math.Ceil(-1 * float64(nElements) * math.Log(targetFP) / (math.Ln2 * math.Ln2)))
	nBlocks := uint32(math.Ceil(float64(bitsNeeded) / float64(blockSize*8)))
	k := uint32(math.Ceil((float64(blockSize*8) / (float64(nElements) / float64(nBlocks))) * math.Ln2))
	totalSize := headerSize + int(nBlocks)*blockSize

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	if err := file.Truncate(int64(totalSize)); err != nil {
		file.Close()
		return nil, err
	}

	data, err := unix.Mmap(int(file.Fd()), 0, totalSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, err
	}

	copy(data[0:4], []byte(magicBytes))
	binary.LittleEndian.PutUint32(data[4:8], fileVersion)
	binary.LittleEndian.PutUint32(data[8:12], k)
	binary.LittleEndian.PutUint32(data[12:16], nBlocks)

	return &BlockedBloomFilter{
		file:    file,
		data:    data,
		k:       k,
		nBlocks: nBlocks,
		hasher:  *NewHasher(),
	}, nil
}

func LoadBlockedBloomFilter(path string) (*BlockedBloomFilter, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	header := make([]byte, headerSize)
	if _, err := file.ReadAt(header, 0); err != nil {
		file.Close()
		return nil, err
	}
	if string(header[0:4]) != magicBytes {
		file.Close()
		return nil, errors.New("invalid bloom filter file")
	}
	k := binary.LittleEndian.Uint32(header[8:12])
	nBlocks := binary.LittleEndian.Uint32(header[12:16])
	totalSize := headerSize + int(nBlocks)*blockSize

	data, err := unix.Mmap(int(file.Fd()), 0, totalSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, err
	}

	return &BlockedBloomFilter{
		file:    file,
		data:    data,
		k:       k,
		nBlocks: nBlocks,
		hasher:  *NewHasher(),
	}, nil
}

func (bf *BlockedBloomFilter) getBlock(h1 uint64) int {
	return int(h1 % uint64(bf.nBlocks))
}

func (bf *BlockedBloomFilter) Add(item []byte) {
	h := NewHasher()
	h1, h2 := h.Sum64(item)

	block := bf.getBlock(h1)
	offset := headerSize + block*blockSize

	for i := uint32(0); i < bf.k; i++ {
		bit := (h1 + uint64(i)*h2) % 512
		byteIndex := int(bit / 8)
		bitOffset := bit % 8
		bf.data[offset+byteIndex] |= 1 << bitOffset
	}
}

func (bf *BlockedBloomFilter) Test(item []byte) bool {
	h := NewHasher()
	h1, h2 := h.Sum64(item)

	block := bf.getBlock(h1)
	offset := headerSize + block*blockSize

	for i := uint32(0); i < bf.k; i++ {
		bit := (h1 + uint64(i)*h2) % 512
		byteIndex := int(bit / 8)
		bitOffset := bit % 8
		if (bf.data[offset+byteIndex] & (1 << bitOffset)) == 0 {
			return false
		}
	}
	return true
}

func (bf *BlockedBloomFilter) Close() error {
	unix.Munmap(bf.data)
	return bf.file.Close()
}
