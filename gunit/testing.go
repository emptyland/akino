package gunit

import (
	"fmt"
	"io"
	"os"
	. "reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/wsxiaoys/terminal/color"
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
	cache    linesCache
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
		cache:    newLinesCache(),
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
		T:         self.t,
		suiteName: self.name,
		name:      caseFn.Name,
		what:      "",
		pwd:       pwd,
		cache:     self.cache,
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
		if throw != nil && (!ok || this != caseOb) {
			fmt.Printf("rethrow %v\n", throw)
			panic(throw)
		}

		if caseOb.numFail > 0 {
			caseOb.T.Fail()
		}
	}()
	if self.setUp != nil {
		self.setUp.Func.Call(args)
	}

	caseFn.Func.Call(args)
	return false
}

type Case struct {
	T           *testing.T
	suiteName   string
	name        string
	what        string
	pwd         string
	numFail     int
	numSucc     int
	breakOnFail bool
	cache       linesCache
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
	if !self.check(2, lhs, rhs, eq) {
		panic(self)
	}
}

func (self *Case) AssertTrue(rhs bool) {
	if !self.check(2, true, rhs, eq) {
		panic(self)
	}
}

func (self *Case) AssertFalse(rhs bool) {
	if !self.check(2, false, rhs, eq) {
		panic(self)
	}
}

func (self *Case) AssertNil(rhs interface{}) {
	if !self.check(2, nil, rhs, eq) {
		panic(self)
	}
}

func (self *Case) AssertNotNil(rhs interface{}) {
	if !self.check(2, nil, rhs, ne) {
		panic(self)
	}
}

func (self *Case) ExpectEquals(lhs, rhs interface{}) {
	self.check(2, lhs, rhs, eq)
}

func (self *Case) ExpectTrue(rhs bool) {
	self.check(2, true, rhs, eq)
}

func (self *Case) ExpectFalse(rhs bool) {
	self.check(2, false, rhs, eq)
}

func (self *Case) ExpectNil(rhs interface{}) {
	self.check(2, nil, rhs, eq)
}

func (self *Case) ExpectNotNil(rhs interface{}) {
	self.check(2, nil, rhs, ne)
}

func (self *Case) String() string {
	return fmt.Sprintf("gunit.Case[%s.%s]", self.suiteName, self.name)
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

func (self *Case) check(depth int, lhs, rhs interface{}, op int) bool {
	defer func() {
		self.what = ""
	}()

	switch op {
	case eq:
		if !DeepEqual(lhs, rhs) {
			self.reportFail(depth+1, lhs, rhs, self.what, op)
			return false
		}
	case ne:
		if DeepEqual(lhs, rhs) {
			self.reportFail(depth+1, lhs, rhs, self.what, op)
			return false
		}
	}
	self.numSucc++
	return true
}

func (self *Case) reportFail(depth int, lhs, rhs interface{}, msg string, op int) {
	self.location(depth + 1)
	fmt.Printf("FAIL (%s) %s\n", opText[op], msg)

	switch op {
	case eq:
		color.Printf("... expected @g%s = %v@|\n", mustTypeName(lhs), lhs)
		color.Printf("...   actual @r%s = %v@|\n", mustTypeName(rhs), rhs)
	case ne:
		color.Printf("... expected @g%s = %v@|\n", mustTypeName(lhs), lhs)
		color.Printf("...   but is @requals@|\n")
	}
	self.numFail++
}

func mustTypeName(ob interface{}) string {
	if ob == nil {
		return "nil"
	} else {
		return TypeOf(ob).Name()
	}
}

func (self *Case) location(depth int) {
	if _, file, line, ok := runtime.Caller(depth); ok {
		lines, err := self.cache.Put(file)
		if err != io.EOF {
			panic(err)
		}

		file = self.fileName(file)
		color.Printf("@y[%s.%s]@| %s:%d:\n%s\n", self.suiteName, self.name, file, line, lines[line-1])
	}
}

func (self *Case) fileName(absPath string) string {
	if strings.HasPrefix(absPath, self.pwd) {
		return strings.TrimPrefix(absPath, self.pwd+"/")
	} else {
		return absPath
	}
}

func (self *Case) Fail() {
	self.location(2)
	fmt.Printf("FAIL %s\n", self.what)
	self.what = ""
	self.numFail++
	panic(self)
}
