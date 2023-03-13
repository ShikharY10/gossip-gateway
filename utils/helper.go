package utils

import (
	crand "crypto/rand"
	"log"
	"math/big"
	"net/http"
	"syscall"
)

func GenerateRandomId() string {
	r32, _ := crand.Int(crand.Reader, big.NewInt(999999999999999))
	p32, _ := crand.Prime(crand.Reader, 5)
	s := r32.String() + p32.String()
	return s
}

// Increase resources limitations
func IncreaseResources() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
}

// Enable pprof hooks
func EnablePProf() {
	if err := http.ListenAndServe("localhost:6060", nil); err != nil {
		log.Fatalf("pprof failed: %v", err)
	}
}
