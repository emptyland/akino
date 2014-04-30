package gunit

import (
	//"fmt"
	"bytes"
	"testing"
)

func TestTesting(t *testing.T) {
	RunTest(&SanityTest{}, t)
}

type SanityTest struct {
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

func (self *SanityTest) TestExpectFail(c *Case) {
	c.ExpectEquals(1, 2)
	c.ExpectEquals(0.1, 0.2)
	c.ExpectEquals("a", "b")
	c.ExpectTrue(false)
	c.ExpectFalse(true)
	c.What("test c must to nil").ExpectNil(c)
	c.ExpectNotNil(c)
	c.ExpectNotNil(nil)
}

func (self *SanityTest) TestBuffer(c *Case) {
	s := `a
b
c
`
	cr := byte('\n')

	buf := bytes.NewBufferString(s)

	l, _ := buf.ReadString(cr)
	c.ExpectEquals("a\n", l)
}

func (self *SanityTest) TestMapRef(c *Case) {
	originMap := map[string]int{
		"a": 0,
		"b": 1,
		"c": 2,
	}
	refMap := originMap
	originMap["b"] = -1
	c.ExpectEquals(-1, refMap["b"])
	c.ExpectEquals(refMap["b"], originMap["b"])
}