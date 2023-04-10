package bytes

// Byte 字节
type Byte int64

// GE 大于等于
func (b Byte) GE(size Byte) bool {
	return b >= size
}

var (
	// B 字节
	B Byte = 1
	// KB 千字节
	KB = Byte(1024) * B
	// MB 兆字节
	MB = Byte(1024) * KB
	// GB 吉字节
	GB = Byte(1024) * MB
	// TB 太字节
	TB = Byte(1024) * GB
)
