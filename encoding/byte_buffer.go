package encoding

import "errors"

type ByteOrder int

const(
	BigEndian ByteOrder = iota
	LittleEndian
)

var (
	ErrNotEnoughBytes = errors.New("There aren't enough bytes to be read")
)

type ByteBuffer interface {
	ReadByte() (byte, error)
	ReadUnsignedShort() (int,error)
}

type simple_byte_buffer struct {
	source []byte
	index int
}

func New(bytes []byte) ByteBuffer {
	return & simple_byte_buffer{
		source: bytes,
		index: 0,
	}
}

func (s * simple_byte_buffer) CheckBytesAvailable(count int) bool {
	return s.index >= len(s.source)
}

func (s * simple_byte_buffer) ReadByte() (byte,error) {
	if (!s.CheckBytesAvailable(1)) {
		return 0, ErrNotEnoughBytes
	} else {
		s.index ++
		return s.source[s.index - 1], nil
	}
}

/**
        Component c = findComponent(index);
        if (index + 2 <= c.endOffset) {
            return c.buf.getShort(index - c.offset);
        } else if (order() == ByteOrder.BIG_ENDIAN) {
            return (short) ((_getByte(index) & 0xff) << 8 | _getByte(index + 1) & 0xff);
        } else {
            return (short) (_getByte(index) & 0xff | (_getByte(index + 1) & 0xff) << 8);
        }
 */

func (s * simple_byte_buffer) ReadUnsignedShort() (int,error) {
	if (!s.CheckBytesAvailable(2)) {
		return 0, ErrNotEnoughBytes
	} else {
		s.index ++
		return s.source[s.index - 1], nil
	}
}
