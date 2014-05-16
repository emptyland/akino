package gunit

import (
	//. "reflect"
	"errors"
)

var errBadArgs = errors.New("gunit: Bad arguments")

// Call the func MUST to be in defer statement.
func (self *Case) ExpectPanic(lhs interface{}) {
	self.expectPanic(3, lhs, recover())
}

func (self *Case) expectPanic(depth int, lhs, rhs interface{}) bool {
	// Assert panic, rethrow it.
	if this, ok := rhs.(*Case); ok {
		panic(this)
	}

	return self.check(depth+1, lhs, rhs, throw)
}

func (self *Case) AssertPanic(lhs, fn func()) {
	defer self.expectPanic(2, lhs, recover())
	fn()
}

func (self *Case) AssertPanicCall(lhs, call interface{}, args ...interface{}) {
	// TODO:
}
