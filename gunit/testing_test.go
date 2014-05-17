package gunit

import (
	"bytes"
	"testing"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
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

func (self *SanityTest) TestLongStringDiff(c *Case) {
	s1 := `this is a
long string.`
	s2 := `that is the
long string`
	c.ExpectEquals(s1, s2)
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

func (self *SanityTest) TestSliceRef(c *Case) {
	originSlice := make([]int, 3, 3)
	refSlice := originSlice

	originSlice[1] = -1
	c.ExpectEquals(-1, originSlice[1])
	c.ExpectEquals(-1, refSlice[1])
}

func (self *SanityTest) TestDiffShow(c *Case) {
	state := dmp.New()
	diffs := state.DiffMain("aaabdef", "aaacdef1", false)

	c.AssertEquals(dmp.Diff{dmp.DiffEqual, "aaa"}, diffs[0])
	c.AssertEquals(dmp.Diff{dmp.DiffDelete, "b"}, diffs[1])
	c.AssertEquals(dmp.Diff{dmp.DiffInsert, "c"}, diffs[2])
	c.AssertEquals(dmp.Diff{dmp.DiffEqual, "def"}, diffs[3])
	c.AssertEquals(dmp.Diff{dmp.DiffInsert, "1"}, diffs[4])
}
