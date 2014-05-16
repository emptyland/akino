package gunit

import (
	"testing"
)

func TestBatchCheck(t *testing.T) {
	RunTest(&BatchCheckTest{}, t)
}

type BatchCheckTest struct {
}

func (self *BatchCheckTest) TestSanity(c *Case) {
	c.Batch(self.foo()).
		AssertEquals(1).
		AssertEquals(-1).
		AssertNil()
}

func (self *BatchCheckTest) TestSkip(c *Case) {
	c.Batch(self.foo()).
		Skip(1).
		AssertEquals(-1).
		AssertNil()
}

func (self *BatchCheckTest) TestExpect(c *Case) {
	c.Batch(self.foo()).
		Skip(1).
		ExpectEquals(-1).
		ExpectNil()

	c.Batch(self.bar()).
		ExpectTrue().
		ExpectFalse()
}

func (self *BatchCheckTest) TestBoolean(c *Case) {
	c.Batch(self.bar()).
		AssertTrue().
		AssertFalse()
}

func (self *BatchCheckTest) foo() (int, int, error) {
	return 1, -1, nil
}

func (self *BatchCheckTest) bar() (bool, bool) {
	return true, false
}
