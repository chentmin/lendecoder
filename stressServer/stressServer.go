package main

import (
	"flag"
	"fmt"
	"github.com/chentmin/lendecoder"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var gomaxprocs *int = flag.Int("p", runtime.NumCPU(), "go max procs")

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var limit = flag.Int("limit", 0, "start limit")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	runtime.GOMAXPROCS(*gomaxprocs)

	addr, err := net.ResolveTCPAddr("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(ln)
	}

	go func() {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				fmt.Println(err)
				continue
			}

			go handleConnection(conn)
		}
	}()

	if *limit > 0 {
		t := time.NewTimer(30e9)
		<-t.C
	} else {
		c := make(chan bool)
		<-c
	}
}

var replyData []byte = []byte{1}

func NewHandler(writer io.Writer) *handler {
	result := new(handler)
	result.writer = writer
	writeChan := make(chan []byte, 100)
	result.writeChan = writeChan
	return result
}

type handler struct {
	writer    io.Writer
	writeChan chan []byte
}

func (h *handler) OnMessage(reader *lendecoder.ReadBuffer) {
	count := 0
	for _, b := range reader.Buffer() {
		count += int(b)
	}
	reply := make([]byte, 2)
	reply[0] = byte(count >> 8)
	reply[1] = byte(count)
	h.writer.Write(reply)
}

func handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			return
		}
	}()

	conn.SetKeepAlive(true)
	conn.SetReadBuffer(4096)
	conn.SetWriteBuffer(4096)
	handler := NewHandler(conn)

	accu := lendecoder.NewAccumulator(handler, 2, 1200)
	err := accu.ReadFrom(conn)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
