package main

import (
	"flag"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	clientCount       = flag.Int("client", 2000, "client count")
	dataSize          = flag.Int64("size", 200, "data size")
	sizeIncludeHeader = flag.Bool("include", true, "header side include header self?")
	destAdd           = flag.String("server", "localhost:8080", "server address")
	gomaxprocs        = flag.Int("p", runtime.NumCPU(), "go max procs")

	data []byte

	currentClientCount = new(int64)

	sg sync.WaitGroup

	totalCount int64

	totalSize int64

	totalLat int64

	statChan = make(chan int64, 30000)
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(*gomaxprocs)

	add, err := net.ResolveTCPAddr("tcp", *destAdd)
	if err != nil {
		fmt.Println(err)
		return
	}

	if *dataSize <= 2 || *dataSize >= 65535 {
		fmt.Printf("data size must be 1 - 65534")
		return
	}

	data = make([]byte, *dataSize)

	sizeToWrite := *dataSize

	if !*sizeIncludeHeader {
		sizeToWrite -= 2
	}

	data[0] = byte(sizeToWrite >> 8)
	data[1] = byte(sizeToWrite)

	go stat()

	c := time.Tick(1e6)
	for i := *clientCount; i > 0; i-- {
		<-c
		sg.Add(1)
		go doConnect(add)
	}

	sg.Wait()
}

func stat() {
	c := time.Tick(2e9)
	var lastTime time.Time
	lastCount := totalCount
	lastSize := totalSize
	lastLat := totalLat
	for {
		select {
		case let := <-statChan:
			totalCount++
			totalSize += *dataSize
			totalLat += let

		case <-c:

			ctime := time.Now()
			thisCount := totalCount
			thisSize := totalSize
			thisLat := totalLat
			duration := ctime.Sub(lastTime).Seconds()
			ops := float64(thisCount-lastCount) / duration
			size := float64(thisSize-lastSize) / duration / 1024 / 1024
			averageLet := float64(thisLat-lastLat) / float64(thisCount-lastCount) / 1e6
			lastTime = ctime
			lastCount = thisCount
			lastSize = thisSize
			lastLat = thisLat
			fmt.Printf("ops: %0.2f, size: %0.2fMB, connection: %d, average latency: %0.2fms\n", ops, size, *currentClientCount, averageLet)

		}
	}
}

func doConnect(add *net.TCPAddr) {
	defer sg.Done()
	conn, err := net.DialTCP("tcp", nil, add)
	if err != nil {
		fmt.Printf("Cannot connect: %v\n", err)
		return
	}
	defer conn.Close()
	*currentClientCount = atomic.AddInt64(currentClientCount, 1)
	defer func() {
		*currentClientCount = atomic.AddInt64(currentClientCount, -1)
	}()

	buf := make([]byte, 10)

	for {
		_, err := conn.Write(data)
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
			return
		}

		sentTime := time.Now()

		_, err = conn.Read(buf)
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
		}

		statChan <- time.Since(sentTime).Nanoseconds()
	}
}
