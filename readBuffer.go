package lendecoder

import (
	"bytes"
)

type ReadBuffer struct {
	data []byte
	*bytes.Reader
}

func NewReadBuffer(buf []byte) *ReadBuffer {
	result := &ReadBuffer{data: buf, Reader: bytes.NewReader(buf)}
	return result
}

/**
return the underlying buffer including the read bytes
*/
func (b *ReadBuffer) Buffer() []byte {
	return b.data
}

func (b *ReadBuffer) ReadableBytes() int {
	return b.Len()
}

func (b *ReadBuffer) ReadUnsignedByte() uint8 {
	a, _ := b.ReadByte()
	return uint8(a)
}

func (b *ReadBuffer) ReadUnsignedShort() uint16 {
	return uint16(b.ReadShort())
}

func (b *ReadBuffer) ReadInt() int32 {
	a1, a2 := b.ReadShort(), b.ReadShort()
	return (int32(a1) << 16) | int32(a2)
}

func (b *ReadBuffer) ReadLong() int64 {
	a1, a2 := b.ReadInt(), b.ReadInt()
	return (int64(a1) << 32) | int64(a2)
}

func (b *ReadBuffer) ReadShort() int16 {
	a, _ := b.ReadByte()
	aa, _ := b.ReadByte()

	return (int16(a) << 8) | int16(aa)
}

type WriteBuffer struct {
	*bytes.Buffer
}

func NewWriteBuffer(size int) *WriteBuffer {
	return &WriteBuffer{bytes.NewBuffer(make([]byte, size))}
}

func (b *WriteBuffer) WriteBytes(data []byte) {
	b.Write(data)
}
