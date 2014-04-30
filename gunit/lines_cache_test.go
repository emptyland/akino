package gunit

import (
	"io"
	"runtime"
	"testing"
)

type LinesCacheTest struct {
	cache linesCache
}

func TestLinesCache(t *testing.T) {
	RunTest(&LinesCacheTest{}, t)
}

func (self *LinesCacheTest) SetUp(c *Case) {
	self.cache = newLinesCache()
	c.AssertNotNil(self.cache)
}

func (self *LinesCacheTest) TearDown(c *Case) {
	c.AssertNotNil(self.cache)
	self.cache = nil
}

func (self *LinesCacheTest) TestSanity(c *Case) {
	var lines []string
	var err error
	var file string
	var line int
	var ok bool

	_, file, line, ok = runtime.Caller(0)
	c.AssertTrue(ok)

	if lines, err = readLines(file); err != io.EOF {
		c.What(err.Error()).Fail()
	}
	c.AssertEquals("	_, file, line, ok = runtime.Caller(0)", lines[line-1])
}
