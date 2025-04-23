package bloom_test

import (
	"blo/bloom"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testBlockedTmpFile = "test_blocked_filter.bloom"
)

func getTestBlockedFilter(size int, fpRate float64) *bloom.BlockedBloomFilter {
	filter, err := bloom.CreateBlockedBloomFilter(testBlockedTmpFile, size, fpRate)
	if err != nil {
		panic(err)
	}

	for i := range size {
		filter.Add([]byte(fmt.Sprintf("id-%d", i)))
	}
	return filter
}

func TestBlockedAddAndTest(t *testing.T) {
	filter := getTestBlockedFilter(testN, testFpRate)
	defer os.Remove(testBlockedTmpFile)

	assert.True(t, filter.Test([]byte("id-42")))
	assert.False(t, filter.Test([]byte("unknown-key")))
}

func TestBlockedFalsePositiveRate(t *testing.T) {
	filter := getTestBlockedFilter(testN, testFpRate)
	defer os.Remove(testBlockedTmpFile)

	fpCount := 0
	total := 50_000
	for i := range total {
		key := fmt.Sprintf("zzz-%d", i)
		if filter.Test([]byte(key)) {
			fpCount++
		}
	}
	fpRate := float64(fpCount) / float64(total)
	t.Logf("False positive rate: %.4f", fpRate)
	assert.LessOrEqual(t, fpRate, 0.02)
}

func TestSlowBlockedFalsePositiveRate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping false positive rate test in short mode")
	}
	targetFpRate := 0.001
	filter := getTestBlockedFilter(500_000_000, targetFpRate)
	defer os.Remove(testBlockedTmpFile)

	fpCount := 0
	total := 50_000
	for i := range total {
		key := fmt.Sprintf("zzz-%d", i)
		if filter.Test([]byte(key)) {
			fpCount++
		}
	}
	fpRate := float64(fpCount) / float64(total)
	t.Logf("False positive rate: %.4f", fpRate)
	assert.LessOrEqual(t, fpRate, targetFpRate)
}

func BenchmarkBlockedBloomMmapTest(b *testing.B) {
	filter := getTestBlockedFilter(testN, testFpRate)
	defer os.Remove(testBlockedTmpFile)

	b.ResetTimer()
	var i uint64 = 0
	for b.Loop() {
		key := []byte(fmt.Sprintf("id-%d", i%testN))
		filter.Test(key)
		i++
	}
}
