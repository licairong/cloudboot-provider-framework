package bytes

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGE(t *testing.T) {
	Convey("字节大于等于比较", t, func() {
		for _, expected := range []struct {
			V1  Byte
			V2  Byte
			Yes bool
		}{
			{V1: B, V2: KB, Yes: false},
			{V1: KB, V2: MB, Yes: false},
			{V1: MB, V2: GB, Yes: false},
			{V1: GB, V2: TB, Yes: false},
			{V1: TB, V2: GB, Yes: true},
			{V1: GB, V2: MB, Yes: true},
			{V1: MB, V2: KB, Yes: true},
			{V1: KB, V2: B, Yes: true},
		} {
			So(expected.V1.GE(expected.V2), ShouldEqual, expected.Yes)
		}
	})
}

func TestByte2MB(t *testing.T) {
	Convey("将字节转化为浮点型的兆字节", t, func() {
		for _, expected := range []struct {
			BValue  Byte
			MBValue string
		}{
			{BValue: Byte(10485760), MBValue: "10.00"},
			{BValue: Byte(1234567), MBValue: "1.18"},
			{BValue: Byte(987654321), MBValue: "941.90"},
		} {
			So(fmt.Sprintf("%.2f", Byte2MB(expected.BValue)), ShouldEqual, expected.MBValue)
		}
	})
}

func TestByte2MBRounding(t *testing.T) {
	Convey("将字节转化为整数的兆字节", t, func() {
		for _, expected := range []struct {
			BValue  Byte
			MBValue int
		}{
			{BValue: Byte(10485760), MBValue: 10},
			{BValue: Byte(1234567), MBValue: 1},
			{BValue: Byte(987654321), MBValue: 941},
		} {
			So(Byte2MBRounding(expected.BValue), ShouldEqual, expected.MBValue)
		}
	})
}

func TestByte2GBRounding(t *testing.T) {
	Convey("将字节转化为整数吉字节", t, func() {
		So(Byte2GBRounding(GB), ShouldEqual, 1)
		So(Byte2GBRounding(7*GB), ShouldEqual, 7)
		So(Byte2GBRounding(1024*MB), ShouldEqual, 1)
		So(Byte2GBRounding(1024*1024*KB), ShouldEqual, 1)
		So(Byte2GBRounding(1024*1024*1024*B), ShouldEqual, 1)
		So(Byte2GBRounding(1023*MB), ShouldEqual, 0)
		So(Byte2GBRounding(1025*MB), ShouldEqual, 1)
		So(Byte2GBRounding(2047*MB), ShouldEqual, 1)
		So(Byte2GBRounding(2048*MB), ShouldEqual, 2)
	})
}

func TestByte2GB(t *testing.T) {
	Convey("将字节转化为吉字节", t, func() {
		So(Byte2GB(GB), ShouldEqual, 1)
		So(Byte2GB(7*GB), ShouldEqual, 7)
		So(Byte2GB(1024*MB), ShouldEqual, 1)
		So(Byte2GB(1024*1024*KB), ShouldEqual, 1)
		So(Byte2GB(1024*1024*1024*B), ShouldEqual, 1)

		So(Byte2GB(1023*MB), ShouldEqual, float64(1023)/float64(1024))
		So(Byte2GB(1025*MB), ShouldEqual, float64(1025)/float64(1024))
		So(Byte2GB(2047*MB), ShouldEqual, float64(2047)/float64(1024))
		So(Byte2GB(2048*MB), ShouldEqual, float64(2048)/float64(1024))
	})
}

func TestParse2Byte(t *testing.T) {
	Convey("容量值解析", t, func() {
		So(TB, ShouldEqual, 1099511627776)
		So(GB, ShouldEqual, 1073741824)
		So(MB, ShouldEqual, 1048576)
		So(KB, ShouldEqual, 1024)
		So(B, ShouldEqual, 1)

		var size Byte
		var err error

		size, err = Parse2Byte("7.5", "MB")
		So(err, ShouldBeNil)
		So(size, ShouldEqual, Byte(7864320))

		size, err = Parse2Byte("837.258", "GB")
		So(err, ShouldBeNil)
		So(size, ShouldEqual, Byte(898998932078))

		size, err = Parse2Byte("7", "AB")
		So(err, ShouldEqual, ErrMalformedUnitStringValue)

		size, err = Parse2Byte("abcdef", "KB")
		So(err, ShouldEqual, ErrMalformedSizeStringValue)

		size, err = Parse2Byte("7", "B")
		So(err, ShouldBeNil)
		So(size, ShouldEqual, Byte(7)*B)

		size, err = Parse2Byte("7", "kb")
		So(err, ShouldBeNil)
		So(size, ShouldEqual, Byte(7)*KB)

		size, err = Parse2Byte("7", "mb")
		So(err, ShouldBeNil)
		So(size, ShouldEqual, Byte(7)*MB)

		size, err = Parse2Byte("7", "GB")
		So(err, ShouldBeNil)
		So(size, ShouldEqual, Byte(7)*GB)

		size, err = Parse2Byte("7", "TB")
		So(err, ShouldBeNil)
		So(size, ShouldEqual, Byte(7)*TB)
	})
}

func TestSplitValueUnit(t *testing.T) {
	Convey("提取容量值与容量单位", t, func() {
		for _, expected := range []struct {
			ValueUnit string
			Value     string
			Unit      string
		}{
			{ValueUnit: "1.2G", Value: "1.2", Unit: "G"},
			{ValueUnit: "5GB", Value: "5", Unit: "GB"},
			{ValueUnit: "5.7 GB", Value: "5.7", Unit: "GB"},
			{ValueUnit: " 5.7 GB ", Value: "5.7", Unit: "GB"},
			{ValueUnit: "KB 5.7", Value: "", Unit: ""},
			{ValueUnit: "5KB 5.7", Value: "5", Unit: "KB 5.7"},
			{ValueUnit: "hello", Value: "", Unit: ""},
		} {
			v, u := SplitValueUnit(expected.ValueUnit)
			So(v, ShouldEqual, expected.Value)
			So(u, ShouldEqual, expected.Unit)
		}
	})
}
