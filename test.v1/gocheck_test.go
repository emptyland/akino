package test

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestGoCheck(t *testing.T) {
	TestingT(t)
}

type TestSuite struct{}

var _ = Suite(&TestSuite{})

func (self *TestSuite) TestHello(c *C) {
	c.Assert(42, Equals, "42")
}
