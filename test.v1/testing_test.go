package test

import (
	//"fmt"
	"testing"
)

func TestTesting(t *testing.T) {
	RunTest(&SanityTest{}, t)
}

type SanityTest struct {
}

func (self *SanityTest) Before() {
}

func (self *SanityTest) After() {
}

func (self *SanityTest) SetUp(c *Case) {
}

func (self *SanityTest) TearDown(c *Case) {
}

func (self *SanityTest) TestSanity(c *Case) {
	c.AssertTrue(true)
	c.AssertFalse(false)
	c.AssertNil(nil)
}

func (self *SanityTest) TestPass(c *Case) {
	c.AssertEquals("ok", "ok")
	c.AssertEquals(1, 1)
}

func (self *SanityTest) TestFail(c *Case) {
	c.What("must fail").AssertEquals(1, "1")
}

func (self *SanityTest) TestOnlyFail(c *Case) {
	c.What("only fail").Fail()
}
