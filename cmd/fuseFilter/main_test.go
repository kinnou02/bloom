package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"unsafe"

	"github.com/FastFilter/xorfilter"
	"github.com/edsrzf/mmap-go"
	"github.com/stretchr/testify/require"
	"github.com/twmb/murmur3"
)

const (
	mainFilterPath = "../../bfuse.dat"
	defaultN       = 1_000_000_000
)

var (
	found bool
	keys  [][]byte
)

func init() {
	keys = make([][]byte, 100_000)
	for i := 0; i < len(keys); i++ {
		keys[i] = []byte(fmt.Sprintf("key-%d", rand.Intn(defaultN)))
	}
}

// ---- mmap loader ----------------------------------------------------------
func loadFuse16(path string) (*xorfilter.BinaryFuse[uint16], mmap.MMap, *os.File, error) {
	// 1. open the file
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, err
	}

	// 2. map it read-only, zero-copy
	mm, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		f.Close()
		return nil, nil, nil, err
	}

	// 3. parse fixed 32-byte header
	const hdrSz = 32
	if len(mm) < hdrSz || string(mm[:4]) != "XF16" {
		mm.Unmap()
		f.Close()
		return nil, nil, nil, fmt.Errorf("invalid header")
	}
	seed := binary.LittleEndian.Uint64(mm[8:])
	segLen := binary.LittleEndian.Uint32(mm[16:])
	segLenMask := binary.LittleEndian.Uint32(mm[20:])
	segCnt := binary.LittleEndian.Uint32(mm[24:])
	segCntLen := binary.LittleEndian.Uint32(mm[28:])

	// 4. wire fingerprints slice directly onto the mmap’d bytes
	fpBytes := mm[hdrSz:]
	fps := unsafe.Slice((*uint16)(unsafe.Pointer(&fpBytes[0])), len(fpBytes)/2)

	filter := &xorfilter.BinaryFuse[uint16]{
		Seed:               seed,
		SegmentLength:      segLen,
		SegmentLengthMask:  segLenMask,
		SegmentCount:       segCnt,
		SegmentCountLength: segCntLen,
		Fingerprints:       fps,
	}

	return filter, mm, f, nil
}

func BenchmarkBloomFromMainFile(b *testing.B) {
	if _, err := os.Stat(mainFilterPath); os.IsNotExist(err) {
		b.Fatalf("File %s not found. Run `go run main.go` first to generate it.", mainFilterPath)
	}

	filter, mm, fi, err := loadFuse16(mainFilterPath)
	require.NoError(b, err)
	defer func() {
		mm.Unmap()
		fi.Close()
	}()

	count := 0
	fmt.Println("Starting benchmark...")
	b.ResetTimer()
	for b.Loop() {
		count++
		key := keys[count%len(keys)]
		found = filter.Contains(murmur3.Sum64(key))
	}
}
