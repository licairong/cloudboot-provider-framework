package bytes

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// Byte2GB 返回字节转化成GB后的数值
func Byte2GB(b Byte) (gb float64) {
	return float64(b) / float64(GB)
}

// Byte2GBRounding 返回字节转化成GB后的整数数值
func Byte2GBRounding(b Byte) (gb int) {
	return int(Byte2GB(b))
}

// Byte2MB 返回字节转化成MB后的数值
func Byte2MB(b Byte) (mb float64) {
	return float64(b) / float64(MB)
}

// Byte2MBRounding 返回字节转化成MB后的整数数值
func Byte2MBRounding(b Byte) (mb int) {
	return int(Byte2MB(b))
}

var (
	// ErrMalformedSizeStringValue 容量值的字符串格式错误
	ErrMalformedSizeStringValue = errors.New("malformed size string value")
	// ErrMalformedUnitStringValue 容量单位的字符串格式错误
	ErrMalformedUnitStringValue = errors.New("malformed unit string value")
)

// Parse2Byte 将字符串类型容量值和容量单位转化成字节
func Parse2Byte(value, unit string) (size Byte, err error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, ErrMalformedSizeStringValue
	}

	switch strings.ToUpper(unit) {
	case "B", "BYTE", "BYTES":
		return B * Byte(val), nil
	case "KB", "K":
		return Byte(int64(1024 * val)), nil
	case "MB", "M":
		return Byte(int64(1024 * 1024 * val)), nil
	case "GB", "G":
		return Byte(int64(1024 * 1024 * 1024 * val)), nil
	case "TB", "T":
		return Byte(int64(1024 * 1024 * 1024 * 1024 * val)), nil
	default:
		return size, ErrMalformedUnitStringValue
	}
}

// SplitValueUnit 从字符串中提取容量值和容量单位。如入参为'1.23GB'，出参为'1.23'和'GB'。
func SplitValueUnit(valueunit string) (value, unit string) {
	valueunit = strings.TrimSpace(valueunit)

	firstLetterIdx := -1
	for i := range valueunit {
		if unicode.IsLetter(rune(valueunit[i])) {
			firstLetterIdx = i
			break
		}
	}
	if firstLetterIdx <= 0 {
		return value, unit
	}
	return strings.TrimSpace(valueunit[:firstLetterIdx]), strings.TrimSpace(valueunit[firstLetterIdx:])
}
