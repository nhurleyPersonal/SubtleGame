package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"subtlegame/server"
	"syscall"
	"time"
)

func main() {
	f, err := os.Create("memprofile.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close()

	// Start memory profiling
	runtime.MemProfileRate = 1
	defer pprof.WriteHeapProfile(f)

	// Handle interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		pprof.WriteHeapProfile(f)
		os.Exit(1)
	}()

	server.Start()

	// Simulate workload
	time.Sleep(30 * time.Second)
}
