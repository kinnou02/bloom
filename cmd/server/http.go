package main

import (
	"blo/bloom"
	"log"
	"net/http"
	_ "net/http/pprof" // for pprof
	"sync"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const filterPath = "blockedbloomfilter.dat"

var (
	filter  *bloom.BlockedBloomFilter
	bufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 256)
		},
	}
)

func main() {
	var err error
	filter, err = bloom.LoadBlockedBloomFilter(filterPath)
	if err != nil {
		log.Fatalf("failed to load bloom filter: %v", err)
	}
	defer filter.Close()

	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing ?key=...", http.StatusBadRequest)
			return
		}

		buf := bufPool.Get().([]byte)
		buf = append(buf[:0], key...) // reset + reuse
		defer bufPool.Put(buf)

		if filter.Test(buf) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("1"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("0"))
		}
	})

	h2s := &http2.Server{}
	serverH2c := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(http.DefaultServeMux, h2s),
	}
	log.Println("Server h2c started on :8080")
	log.Fatal(serverH2c.ListenAndServe())
}
