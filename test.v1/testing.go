package test

import (
	"fmt"
	"os"
	. "reflect"
	"runtime"
	"strings"
	"testing"
)

//
// type Test struct {
// }
//
// func (self *Test) Before()
//
// func (self *Test) SetUp(c *Case)
//
// func (self *Test) TestXXX()
//
// func (self *Test) TearDown(c *Case)
//
// func (self *Test) After()
//
const (
	beforeName   = "Before"
	setUpName    = "SetUp"
	testPrefix   = "Test"
	tearDownName = "TearDown"
	afterName    = "After"
)

type Test struct {
	unit     interface{}
	unitType Type
	t        *testing.T
	name     string
	before   *Method
	after    *Method
	setUp    *Method
	tearDown *Method
	numSucc  int
	numFail  int
}

func RunTest(unit interface{}, t *testing.T) {
	unitType := TypeOf(unit)
	name := unitType.Name()
	if unitType.Kind() == Ptr {
		name = Indirect(ValueOf(unit)).Type().Name()
	}
	suite := &Test{
		unit:     unit,
		unitType: unitType,
		t:        t,
		name:     name,
		before:   method(unitType, beforeName),
		after:    method(unitType, afterName),
		setUp:    method(unitType, setUpName),
		tearDown: method(unitType, tearDownName),
	}
	suite.Run()
}

func method(unitType Type, name string) *Method {
	if found, ok := unitType.MethodByName(name); ok {
		return &found
	} else {
		return nil
	}
}

func (self *Test) Run() {
	if self.before != nil {
		self.before.Func.Call([]Value{ValueOf(self.unit)})
	}
	defer func() {
		if self.after != nil {
			self.after.Func.Call([]Value{ValueOf(self.unit)})
		}
	}()

	for i := 0; i < self.unitType.NumMethod(); i++ {
		caseFn := self.unitType.Method(i)
		if strings.HasPrefix(caseFn.Name, testPrefix) {
			self.runCase(caseFn)
		}
	}
}

func (self *Test) runCase(caseFn Method) bool {
	pwd, err := os.Getwd()
	if err != nil {
		return true
	}
	caseOb := &Case{
		t:         self.t,
		suiteName: self.name,
		name:      caseFn.Name,
		what:      "",
		pwd:       pwd,
	}
	args := []Value{
		ValueOf(self.unit),
		ValueOf(caseOb),
	}

	defer func() {
		if self.tearDown != nil {
			self.tearDown.Func.Call(args)
		}

		throw := recover()
		this, ok := throw.(*Case)
		if !ok || this != caseOb {
			panic(throw)
		}
	}()
	if self.setUp != nil {
		self.setUp.Func.Call(args)
	}

	caseFn.Func.Call(args)
	return false
}

type Case struct {
	t         *testing.T
	suiteName string
	name      string
	what      string
	pwd       string
}

func (self *Case) Run(fn func(c *Case)) (fail bool) {
	defer func() {
		throw := recover()
		if throw != nil {
			fail = true
		}
		this, ok := throw.(*Case)
		if !ok || this != self {
			panic(throw)
		}
	}()
	fn(self)
	return false
}

func (self *Case) What(format string, args ...interface{}) *Case {
	self.what = fmt.Sprintf(format, args...)
	return self
}

func (self *Case) AssertEquals(lhs, rhs interface{}) {
	self.assertEquals(2, lhs, rhs)
}

func (self *Case) AssertTrue(rhs bool) {
	self.assertEquals(2, true, rhs)
}

func (self *Case) AssertFalse(rhs bool) {
	self.assertEquals(2, false, rhs)
}

func (self *Case) AssertNil(rhs interface{}) {
	self.assertEquals(2, nil, rhs)
}

const (
	eq = iota
	ne
	lt
	le
	gt
	ge
)

var opText = []string{
	"==",
	"!=",
	"<",
	"<=",
	">",
	">=",
}

func (self *Case) assertEquals(depth int, lhs, rhs interface{}) {
	defer func() {
		self.what = ""
	}()

	self.assert(depth+1, lhs, rhs, self.what, eq)
}

func (self *Case) assert(depth int, lhs, rhs interface{}, msg string, op int) {
	switch op {
	case eq:
		if !DeepEqual(lhs, rhs) {
			self.raiseFail(depth+1, lhs, rhs, msg, op)
		}
	case ne:
		if DeepEqual(lhs, rhs) {
			self.raiseFail(depth+1, lhs, rhs, msg, op)
		}
	}
}

func (self *Case) raiseFail(depth int, lhs, rhs interface{}, msg string, op int) {
	self.location(depth + 1)
	fmt.Printf("    assert fail! (%s) %s\n", opText[op], msg)
	fmt.Printf("... expected %s = %v\n", TypeOf(lhs).Name(), lhs)
	fmt.Printf("...   actual %s = %v\n", TypeOf(rhs).Name(), rhs)
	panic(self)
}

func (self *Case) location(depth int) {
	if _, file, line, ok := runtime.Caller(depth); ok {
		file = self.fileName(file)
		fmt.Printf("[%s.%s] %s:%d:\n", self.suiteName, self.name, file, line)
	}
}

func (self *Case) fileName(absPath string) string {
	if strings.HasPrefix(absPath, self.pwd) {
		return strings.TrimPrefix(absPath, self.pwd)
	} else {
		return absPath
	}
}

func (self *Case) Fail() {
	self.location(2)
	fmt.Printf("    fail! %s\n", self.what)
	self.what = ""
	panic(self)
}
