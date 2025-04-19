// bloom/bloom.go
package bloom

import (
	"encoding/binary"
	"errors"
	"os"

	"github.com/cespare/xxhash/v2"
	"golang.org/x/exp/mmap"
)

type BloomFilter struct {
	Bitset []byte
	M      uint
	K      uint
}

func New(m, k uint) *BloomFilter {
	return &BloomFilter{
		Bitset: make([]byte, (m+7)/8),
		M:      m,
		K:      k,
	}
}

// double hashing: h_i(x) = h1(x) + i * h2(x) % m
func (bf *BloomFilter) hashes(data []byte) []uint {
	h1 := xxhash.Sum64(data)
	h2 := xxhash.Sum64(append([]byte{0}, data...))

	hashes := make([]uint, bf.K)
	for i := uint(0); i < bf.K; i++ {
		hashes[i] = uint((h1 + uint64(i)*h2) % uint64(bf.M))
	}
	return hashes
}

func (bf *BloomFilter) Add(data []byte) {
	for _, h := range bf.hashes(data) {
		byteIdx, bitIdx := h/8, h%8
		bf.Bitset[byteIdx] |= 1 << bitIdx
	}
}

func (bf *BloomFilter) Test(data []byte) bool {
	for _, h := range bf.hashes(data) {
		byteIdx, bitIdx := h/8, h%8
		if (bf.Bitset[byteIdx] & (1 << bitIdx)) == 0 {
			return false
		}
	}
	return true
}

const magic = "BLOOMv1"

func SaveToFile(path string, bf *BloomFilter) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write([]byte(magic)); err != nil {
		return err
	}
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[0:8], uint64(bf.M))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(bf.K))
	if _, err := f.Write(buf); err != nil {
		return err
	}
	_, err = f.Write(bf.Bitset)
	return err
}

func LoadFromFile(path string) (*BloomFilter, func(), error) {
	reader, err := mmap.Open(path)
	if err != nil {
		return nil, nil, err
	}

	header := make([]byte, len(magic))
	_, err = reader.ReadAt(header, 0)
	if err != nil || string(header) != magic {
		return nil, nil, errors.New("invalid bloom file header")
	}

	meta := make([]byte, 16)
	_, err = reader.ReadAt(meta, int64(len(magic)))
	if err != nil {
		return nil, nil, err
	}
	m := binary.LittleEndian.Uint64(meta[0:8])
	k := binary.LittleEndian.Uint64(meta[8:16])

	offset := int64(len(magic) + 16)
	size := reader.Len() - int(offset)
	data := make([]byte, size)
	_, err = reader.ReadAt(data, offset)
	if err != nil {
		return nil, nil, err
	}

	bf := &BloomFilter{
		M:      uint(m),
		K:      uint(k),
		Bitset: data,
	}

	cleanup := func() { reader.Close() }
	return bf, cleanup, nil
}
