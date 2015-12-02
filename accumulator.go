package lendecoder

import (
	"errors"
	"io"
)

const (
	defaultBufferSize = 4096
)

var (
	ErrTooLongFrame = errors.New("Error frame too long")
)

type MessageHandler interface {
	OnMessage(*ReadBuffer)
}

type Accumulator struct {
	buf               []byte
	readerIndex       int
	writerIndex       int
	handler           MessageHandler
	lengthFieldLength int
	maxPacketLength   int
}

func NewAccumulator(handler MessageHandler, lengthFieldLength, maxPacketLength int) *Accumulator {
	result := &Accumulator{buf: make([]byte, defaultBufferSize), handler: handler, lengthFieldLength: lengthFieldLength, maxPacketLength: maxPacketLength}
	return result
}

func (accu *Accumulator) ReadFrom(reader io.ReadCloser) error {
	defer reader.Close()
	for {
		n, err := reader.Read(accu.buf[accu.writerIndex:])
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if n == 0 {
			continue
		}

		accu.writerIndex += n

		for readableBytes := accu.ReadableBytes(); readableBytes >= accu.lengthFieldLength; {

			var frameLength int
			switch accu.lengthFieldLength {
			case 2:
				frameLength = accu.getShort(accu.readerIndex)
			default:
				break
			}

			if frameLength > accu.maxPacketLength {
				return ErrTooLongFrame
			}

			if readableBytes >= frameLength {
				accu.handler.OnMessage(NewReadBuffer(accu.buf[accu.readerIndex+accu.lengthFieldLength : accu.readerIndex+frameLength]))
				accu.readerIndex += frameLength
				readableBytes -= frameLength
			} else {
				break
			}
		}

		if len(accu.buf)-accu.writerIndex <= 100 {
			// new buffer
			readableBytes := accu.ReadableBytes()
			if readableBytes == 0 {
				accu.buf = make([]byte, defaultBufferSize)
				accu.readerIndex = 0
				accu.writerIndex = 0
			} else {
				newLength := readableBytes + defaultBufferSize
				newBuf := make([]byte, newLength)
				copy(newBuf, accu.buf[accu.readerIndex:accu.writerIndex])
				accu.buf = newBuf
				accu.readerIndex = 0
				accu.writerIndex = readableBytes
			}
		}
	}
	return nil
}

func (accu *Accumulator) getShort(index int) int {
	return (int(accu.buf[index]) << 8) | int(accu.buf[index+1])
}

func (accu *Accumulator) Readable() bool {
	return accu.ReadableBytes() > 0
}

func (accu *Accumulator) ReadableBytes() int {
	return accu.writerIndex - accu.readerIndex
}
