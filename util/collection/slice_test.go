package collection

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInSlice(t *testing.T) {
	Convey("测试目标字符串是否在字符串集合中", t, func() {
		for _, expected := range []struct {
			Slice    []string
			V        string
			Contains bool
		}{
			{Slice: []string{"a", "b", "c"}, V: "b", Contains: true},
			{Slice: []string{"a", "b", "c"}, V: "d", Contains: false},
			{Slice: []string{"a", "b", "c"}, V: " a", Contains: false},
		} {
			So(InSlice(expected.V, expected.Slice), ShouldEqual, expected.Contains)
		}
	})
}
