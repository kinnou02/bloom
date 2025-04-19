package bloom_test

import (
	"bloomfilter/bloom"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testN       = 100_000
	testFpRate  = 0.01
	testTmpFile = "test_filter.bloom"
)

func getTestFilter(t *testing.T) *bloom.BloomFilter {
	m := uint(math.Ceil(-1 * float64(testN) * math.Log(testFpRate) / (math.Ln2 * math.Ln2)))
	k := uint(math.Ceil((float64(m) / float64(testN)) * math.Ln2))
	filter := bloom.New(m, k)

	for i := 0; i < testN; i++ {
		filter.Add([]byte(fmt.Sprintf("id-%d", i)))
	}
	return filter
}

func TestAddAndTest(t *testing.T) {
	filter := getTestFilter(t)

	assert.True(t, filter.Test([]byte("id-42")))
	assert.False(t, filter.Test([]byte("unknown-key")))
}

func TestSaveAndLoad(t *testing.T) {
	filter := getTestFilter(t)

	err := bloom.SaveToFile(testTmpFile, filter)
	assert.NoError(t, err)
	defer os.Remove(testTmpFile)

	loaded, cleanup, err := bloom.LoadFromFile(testTmpFile)
	assert.NoError(t, err)
	defer cleanup()

	assert.True(t, loaded.Test([]byte("id-42")))
	assert.False(t, loaded.Test([]byte("unknown-key")))
}

func TestFalsePositiveRate(t *testing.T) {
	filter := getTestFilter(t)

	fpCount := 0
	total := 50_000
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("zzz-%d", i)
		if filter.Test([]byte(key)) {
			fpCount++
		}
	}
	fpRate := float64(fpCount) / float64(total)
	t.Logf("False positive rate: %.4f", fpRate)
	assert.LessOrEqual(t, fpRate, 0.02) // tolère un petit dépassement
}

func BenchmarkBloomMmapTest(b *testing.B) {
	filter := getTestFilter(b)

	// Save then load using mmap
	tmp := filepath.Join(os.TempDir(), "bench_filter.bloom")
	err := bloom.SaveToFile(tmp, filter)
	assert.NoError(b, err)
	defer os.Remove(tmp)

	mapped, cleanup, err := bloom.LoadFromFile(tmp)
	assert.NoError(b, err)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("id-%d", i%testN))
		mapped.Test(key)
	}
}
