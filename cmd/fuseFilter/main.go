package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"

	// adapte ce chemin à ton module si besoin

	"github.com/FastFilter/xorfilter"
	"github.com/twmb/murmur3"
)

const (
	numElements  = 500_000_000
	falsePosRate = 0.0001
	filterFile   = "fusefilter.dat"
)

// 32-byte on-disk header, little-endian
type header struct {
	Magic              [8]byte // e.g. {'X','F','1','6',0,0,0,0}
	Seed               uint64
	SegmentLength      uint32
	SegmentLengthMask  uint32
	SegmentCount       uint32
	SegmentCountLength uint32
}

func main() {
	fmt.Println("constructing elements...")
	iElems := make([]uint64, numElements)
	for i := 0; i < numElements; i++ {
		v := []byte(fmt.Sprintf("key-%d", i))
		iElems[i] = murmur3.Sum64(v)
	}
	fmt.Println("creating binaryFuse...")
	filter, err := xorfilter.NewBinaryFuse[uint16](iElems)
	if err != nil {
		panic(err)
	}

	fp16 := filter.Fingerprints
	if len(fp16) == 0 {
		return // rien à écrire
	}
	f, err := os.Create("bfuse.dat")
	if err != nil {
		panic(err)
	}
	binary.Write(f, binary.LittleEndian, header{
		Magic: [8]byte{'X', 'F', '1', '6', 0, 0, 0, 0},
		Seed:  filter.Seed, SegmentLength: filter.SegmentLength,
		SegmentLengthMask:  filter.SegmentLengthMask,
		SegmentCount:       filter.SegmentCount,
		SegmentCountLength: filter.SegmentCountLength,
	})
	binary.Write(f, binary.LittleEndian, filter.Fingerprints) // []uint16
	f.Close()
	fmt.Printf("Saved filter to %s\n", filterFile)

	// Vérif d'un élément connu
	testKey := []byte("key-42")
	fmt.Printf("Test key-42: %v (should be true)\n", filter.Contains(murmur3.Sum64(testKey)))

	// Faux positifs
	fmt.Println("Testing false positives...")
	rand.Seed(time.Now().UnixNano())
	falsePos := 0
	total := 1_000_000
	for i := 0; i < total; i++ {
		k := fmt.Sprintf("zzz-%d", rand.Int())
		if filter.Contains(murmur3.Sum64([]byte(k))) {
			falsePos++
		}
	}
	fmt.Printf("False positives: %d / %d (%.6f)\n", falsePos, total, float64(falsePos)/float64(total))
}
