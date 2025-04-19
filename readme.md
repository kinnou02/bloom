bloom filter backed on disc.
memory is use for creation, then it is written to disk.
this bloom filter can be loaded in readonly to check if an element is in it or not.

## Perf
```
goos: linux
goarch: amd64
pkg: blo
cpu: Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz
BenchmarkBloomFromMainFile-8            22339610               513.8 ns/op           120 B/op        4 allocs/op
BenchmarkBloomParallelTest-8            125923819              102.9 ns/op           120 B/op        4 allocs/op
BenchmarkBitsetAndBloomMemory-8         22973782               537.4 ns/op            24 B/op        2 allocs/op
```
test are run with a 500 millions values bloom filter and a false positive rate of 0.001

## benchmark HTTP
```
wrk -t 6 -c100  -d60s  -s bloom.lua 'http://localhost:8080/'
```
