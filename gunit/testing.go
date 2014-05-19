package gunit

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	. "reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
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

var testName = flag.String("gunit.test", "", "specify the name for test running.")
var testV = flag.Bool("gunit.v", false, "enable verbose output.")

var specSuiteName = ""
var specCaseName = ""

func init() {
	flag.Parse()

	if testName == nil || *testName == "" {
		return
	}

	par := strings.Split(*testName, ".")
	if len(par) <= 1 {
		specSuiteName = *testName
	} else {
		specSuiteName = par[0]
		specCaseName = par[1]
	}
}

func RunTest(unit interface{}, t *testing.T) {
	unitType := TypeOf(unit)
	name := unitType.Name()
	if unitType.Kind() == Ptr {
		name = Indirect(ValueOf(unit)).Type().Name()
	}

	if specSuiteName != "" && specSuiteName != name {
		t.SkipNow()
		return // Skip no specity suite
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
		if specCaseName != "" && specCaseName != caseFn.Name {
			continue
		}

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

		origin := recover()
		this, ok := origin.(*Case)
		if origin != nil && (!ok || this != caseOb) {
			panic(origin)
		}

		if caseOb.numFail > 0 {
			caseOb.T.Fail()
		}
	}()
	if self.setUp != nil {
		self.setUp.Func.Call(args)
	}

	var jiffy time.Time
	if *testV {
		jiffy = time.Now()
	}
	caseFn.Func.Call(args)
	if *testV {
		caseOb.reportPass(jiffy)
	}
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
		origin := recover()
		if origin != nil {
			fail = true
		}
		this, ok := origin.(*Case)
		if !ok || this != self {
			panic(origin)
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
	throw
)

var opText = []string{
	"==",
	"!=",
	"<",
	"<=",
	">",
	">=",
	"panic",
}

func (self *Case) check(depth int, lhs, rhs interface{}, op int) bool {
	defer func() {
		self.what = ""
	}()

	switch op {
	case eq:
		fallthrough
	case throw:
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

func (self *Case) reportPass(jiffy time.Time) {
	color.Printf("@g---@| %s.%s\n@g---@| PASS %0.3fs\n", self.suiteName, self.name, time.Since(jiffy).Seconds())
}

func (self *Case) reportFail(depth int, lhs, rhs interface{}, msg string, op int) {
	self.location(depth + 1)
	color.Printf("@r---@| FAIL %s\n", msg)

	switch op {
	case eq:
		fallthrough
	case throw:
		self.reportFailWithDiff(lhs, rhs)
	case ne:
		color.Printf("@r...@| expected @g%s = %v@|\n", mustTypeName(lhs), lhs)
		color.Printf("@r...@|   but is @requals@|\n")
	}
	self.numFail++
}

func (self *Case) reportFailWithDiff(lhs, rhs interface{}) {
	if lhs != nil && rhs != nil {
		if s1, ok := lhs.(string); ok {
			s2 := rhs.(string)
			self.printDiff(s1, s2)
			return
		}
	}

	color.Printf("@r...@| expected @g%s = %v@|\n", mustTypeName(lhs), lhs)
	color.Printf("@r...@|   actual @r%s = %v@|\n", mustTypeName(rhs), rhs)
}

func (self *Case) printDiff(s1, s2 string) {
	s1, s2 = formatize("    ", s1), formatize("    ", s2)
	color.Printf("@r...@| expected @gstring =@|\n%v\n", s1)
	color.Printf("@r...@|   actual @rstring =@|\n%v\n", s2)
	color.Printf("@r...@|  diff is\n")

	state := dmp.New()
	diffs := state.DiffMain(s1, s2, false)
	for _, diff := range diffs {
		switch diff.Type {
		case dmp.DiffEqual:
			fmt.Print(diff.Text)

		case dmp.DiffDelete:
			color.Printf("@r[-%s]@|", diff.Text)

		case dmp.DiffInsert:
			color.Printf("@g[+%s]@|", diff.Text)
		}
	}
	fmt.Println()
}

func formatize(indent, s string) string {
	lines := strings.Split(s, "\n")

	var buf bytes.Buffer
	for i, line := range lines {
		if i > 0 {
			buf.WriteRune('\n')
		}
		buf.WriteString(indent)
		buf.WriteString(line)
	}
	return buf.String()
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
		color.Printf("@r---@| %s.%s %s:%d:\n%s\n", self.suiteName, self.name, file, line, lines[line-1])
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
	color.Printf("@r---@| FAIL %s\n", self.what)
	self.what = ""
	self.numFail++
	panic(self)
}
