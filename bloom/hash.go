package bloom

import (
	"hash/maphash"

	"github.com/cespare/xxhash/v2"
	"github.com/spaolacci/murmur3"
)

type Hasher interface {
	Sum64([]byte) (uint64, uint64)
}

func NewHasher() Hasher {
	return &MixedHasher{}
}

type XXHasher struct{}

func (h *XXHasher) Sum64(data []byte) (uint64, uint64) {
	h1 := xxhash.Sum64(data)

	salted := append(data, 'X') // simple dérivation
	h2 := xxhash.Sum64(salted)

	return h1, h2
}

// sum64 function using mumur3
type MumurHasher struct{}

func (h *MumurHasher) Sum64(data []byte) (uint64, uint64) {
	h1 := murmur3.Sum64(data)
	salted := append(data, 'X') // simple dérivation
	h2 := murmur3.Sum64(salted)
	return h1, h2
}

type MixedHasher struct{}

func (h *MixedHasher) Sum64(data []byte) (uint64, uint64) {
	h1 := xxhash.Sum64(data)
	h2 := murmur3.Sum64(data)
	return h1, h2
}

type MapHasher struct{}

func (h *MapHasher) Sum64(data []byte) (uint64, uint64) {
	mh := maphash.Hash{}
	mh.Write(data)
	h1 := mh.Sum64()
	mh.Write([]byte("X")) // simple dérivation
	h2 := mh.Sum64()
	return h1, h2
}
