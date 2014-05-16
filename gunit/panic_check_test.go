package gunit

import (
	"testing"
)

type PanicCheckTest struct {
}

func TestPanicCheck(t *testing.T) {
	RunTest(&PanicCheckTest{}, t)
}

func (self *PanicCheckTest) TestSanity(c *Case) {
	defer c.ExpectPanic(1)

	panic(1)
}

func (self *PanicCheckTest) TestFail(c *Case) {
	defer c.ExpectPanic("ok")

	panic("fail")
}
