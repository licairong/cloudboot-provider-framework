package collection

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSSet(t *testing.T) {
	Convey("字符串集合", t, func() {
		Convey("删除", func() {
			set := NewSSet(-3, "a", "b", "c").Add("d", "e", "f").Remove("b", "a", "abc")
			So(set.Length(), ShouldEqual, 4)
			So(set.Contains("c"), ShouldBeTrue)
			So(set.Elements(), ShouldContain, "d")
			So(set.Elements(), ShouldContain, "e")
			So(set.Elements(), ShouldContain, "f")
			So(set.IsEmpty(), ShouldBeFalse)
		})
	})
}
