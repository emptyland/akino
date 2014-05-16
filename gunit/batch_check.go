package gunit

import (
	"fmt"
)

// Example:
// func Foo() (int, int, string, error)
//
// c.Batch(Foo()).
// 	AssertEquals(1).
// 	Skip(1).
// 	AssertEquals("ok").
// 	AssertNil()
//

type Batch struct {
	c          *Case
	returnVals []interface{}
	index      int
}

// Accept all func return values for check.
func (self *Case) Batch(returnVals ...interface{}) *Batch {
	return &Batch{
		c:          self,
		returnVals: returnVals,
	}
}

func (self *Batch) Skip(offset int) *Batch {
	self.index += offset
	return self
}

func (self *Batch) AssertNil() *Batch {
	if !self.c.check(2, nil, self.val(), eq) {
		panic(self.c)
	}
	return self.Skip(1)
}

func (self *Batch) AssertNotNil() *Batch {
	if !self.c.check(2, nil, self.val(), ne) {
		panic(self.c)
	}
	return self.Skip(1)
}

func (self *Batch) AssertTrue() *Batch {
	if !self.c.check(2, true, self.val(), eq) {
		panic(self.c)
	}
	return self.Skip(1)
}

func (self *Batch) AssertFalse() *Batch {
	if !self.c.check(2, false, self.val(), eq) {
		panic(self.c)
	}
	return self.Skip(1)
}

func (self *Batch) AssertEquals(lhs interface{}) *Batch {
	if !self.c.check(2, lhs, self.val(), eq) {
		panic(self.c)
	}
	return self.Skip(1)
}

func (self *Batch) ExpectNil() *Batch {
	self.c.check(2, nil, self.val(), eq)
	return self.Skip(1)
}

func (self *Batch) ExpectNotNil() *Batch {
	self.c.check(2, nil, self.val(), ne)
	return self.Skip(1)
}

func (self *Batch) ExpectTrue() *Batch {
	self.c.check(2, true, self.val(), eq)
	return self.Skip(1)
}

func (self *Batch) ExpectFalse() *Batch {
	self.c.check(2, false, self.val(), eq)
	return self.Skip(1)
}

func (self *Batch) ExpectEquals(lhs interface{}) *Batch {
	self.c.check(2, lhs, self.val(), eq)
	return self.Skip(1)
}

func (self *Batch) val() interface{} {
	if self.index >= len(self.returnVals) {
		self.c.location(1)
		fmt.Printf("FAIL: too many check values, max: %v\n", len(self.returnVals))
		self.c.numFail++
		panic(self.c)
	}
	return self.returnVals[self.index]
}
