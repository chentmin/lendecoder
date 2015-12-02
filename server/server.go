package main

import (
	"fmt"
	"github.com/chentmin/flashsecurity"
	"github.com/chentmin/lendecoder"
	"io"
	"net"
)

func main() {
	go func() {
		err := flashsecurity.Serve()
		if err != nil {
			fmt.Printf("Cannot start security server: %v\n", err)
		}
	}()
	addr, err := net.ResolveTCPAddr("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(ln)
	}

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

type handler struct {
	writer io.Writer
}

func (h *handler) OnConnected(writer io.Writer) {
	h.writer = writer
}

func (h *handler) OnMessage(reader *lendecoder.ReadBuffer) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			return
		}
	}()
	checkSum := reader.ReadUnsignedByte()
	fmt.Printf("CheckSum: %v\n", checkSum)

	bigOffset := reader.ReadUnsignedByte()
	fmt.Printf("big offset: %v\n", bigOffset)

	msgId := reader.ReadUnsignedShort()
	fmt.Printf("offset: %v\n", msgId>>13)
	fmt.Printf("Received msg %v\n", msgId&0x1fff)
}

func (h *handler) OnDisconnected() {

	fmt.Println("client disconnected")
}

func NewHandler() *handler {
	return new(handler)
}

func handleConnection(conn *net.TCPConn) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			return
		}
	}()

	handler := NewHandler()
	handler.OnConnected(conn)
	defer handler.OnDisconnected()

	accu := lendecoder.NewAccumulator(handler, 2, 1200)
	err := accu.ReadFrom(conn)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
