package gproxy

import (
	"bytes"
	"io"
	"reflect"
	"unsafe"
)

type BufferHelper struct {
	Reader *bytes.Buffer
}

func (p *BufferHelper) Read(val []byte) (n int, err error) {
	if p.Reader.Len() == 0 {
		p.resetOffset()

		return 0, io.EOF
	}

	n, err = p.Reader.Read(val)

	return
}

func (p *BufferHelper) resetOffset() {
	r := reflect.ValueOf(p.Reader)
	buffer := r.Elem()

	offValue := buffer.FieldByName("off")
	offValue = reflect.NewAt(offValue.Type(), unsafe.Pointer(offValue.UnsafeAddr())).Elem()
	offValue.SetInt(0)

	lastReadValue := buffer.FieldByName("lastRead")
	lastReadValue = reflect.NewAt(lastReadValue.Type(), unsafe.Pointer(lastReadValue.UnsafeAddr())).Elem()
	lastReadValue.SetInt(0)
}

func (p *BufferHelper) Close() error {
	return nil
}

func ReaderCloser(r io.Reader) *BufferHelper {

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return &BufferHelper{Reader: buf}
}
