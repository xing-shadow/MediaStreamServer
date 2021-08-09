package RichConn

import (
	"bufio"
	"io"
)

type ReaderWriter struct {
	*bufio.ReadWriter
	readErr  error
	writeErr error
}

func NewReadrWriter(rw io.ReadWriter, bufSize int) *ReaderWriter {
	return &ReaderWriter{
		ReadWriter: bufio.NewReadWriter(bufio.NewReaderSize(rw, bufSize), bufio.NewWriterSize(rw, bufSize)),
	}
}

func (pThis *ReaderWriter) Read(p []byte) (int, error) {
	if pThis.readErr != nil {
		return 0, pThis.readErr
	}
	n, err := io.ReadAtLeast(pThis, p, len(p))
	pThis.readErr = err
	return n, err
}

func (pThis *ReaderWriter) ReadUintBE(n int) (uint32, error) {
	if pThis.readErr != nil {
		return 0, pThis.readErr
	}
	ret := uint32(0)
	for i := 0; i < n; i++ {
		b, err := pThis.Reader.ReadByte()
		if err != nil {
			pThis.readErr = err
			return 0, err
		}
		ret = ret<<8 + uint32(b)
	}
	return ret, nil
}

func (pThis *ReaderWriter) ReadUintLE(n int) (uint32, error) {
	if pThis.readErr != nil {
		return 0, pThis.readErr
	}
	ret := uint32(0)
	for i := 0; i < n; i++ {
		b, err := pThis.Reader.ReadByte()
		if err != nil {
			pThis.readErr = err
			return 0, err
		}
		ret += uint32(b) << uint32(8*i)
	}
	return ret, nil
}

func (pThis *ReaderWriter) Write(p []byte) (int, error) {
	if pThis.writeErr != nil {
		return 0, pThis.writeErr
	}
	return pThis.Writer.Write(p)
}

func (pThis *ReaderWriter) WriteUintBE(v uint32, n int) error {
	if pThis.writeErr != nil {
		return pThis.writeErr
	}
	for i := 0; i < n; i++ {
		b := byte(v>>uint32((n-i-1)*8)) & 0xff
		if err := pThis.Writer.WriteByte(b); err != nil {
			pThis.writeErr = err
			return err
		}
	}
	return nil
}

func (pThis *ReaderWriter) WriteUintLE(v uint32, n int) error {
	if pThis.writeErr != nil {
		return pThis.writeErr
	}
	for i := 0; i < n; i++ {
		b := byte(v) & 0xff
		if err := pThis.Writer.WriteByte(b); err != nil {
			pThis.writeErr = err
			return err
		}
		v = v >> 8
	}
	return nil
}

func (pThis *ReaderWriter) Flush() error {
	if pThis.writeErr != nil {
		return pThis.writeErr
	}

	if pThis.Writer.Buffered() == 0 {
		return nil
	}
	return pThis.Writer.Flush()
}
