# Gunit - Go Unit Test

@(Gunit Document)[go|design]

gunit add rich assert and expect checker for `go test`.

## Install

```bash
go get github.com/emptyland/akino/gunit
```

## Add a Test Suite

```go
package foo

import(
    "testing"
    
    "github.com/emptyland/akino/gunit"
)

// Your suite struct
type MySuite struct {
    a int
    b int
    t string
}

// Suite entry
func Test(t *testing.T) {
    gunit.RunTest(&MySuite{}, t)
}

// Optional: before suite
func (self *MySuite) Before() {
}

// Optional: after suite
func (self *MySuite) After() {
}
```

This is a simple test suite for gunit.

Test source files like `go test` file, it must be named `xxxx_test.go`

### How to run it?

Run all test:
```
go test
```

Run one test suite:
```
go test -gunit.test=<suiteName>
```

Run specified test case:
```
go test -gunit.test=<suiteName.caseName>
```

The order of suite running:
```
-- suite scope
suite.Before()

    -- test 1 scope
    suite.SetUp()
    suite.TestXXX1
    suite.TearDown()
    -- test 1 end

    -- test 2 scope
    suite.SetUp()
    suite.TestXXX2
    suite.TearDown()
    -- test 2 end

suite.After()
-- suite end
```

> In a suite: Before()/After() run only once.

You should add `case` to your suite for testing.

## Add Test Cases

```go
// Optional: test case setup
func (self *MySuite) SetUp(c *gunit.Case) {
    self.a = 314
    self.b = 216
    self.t = "hello"
}

// Optional: test case teardown
func (self *MySuite) TearDown(c *gunit.Case) {
    *self = MySuite{} // clear the suite
}

// Test case must be prefix with Test.
func (self *MySuite) TestSanity(c *gunit.Case) {
    c.AssertEquals(314, self.a)
    c.AssertEquals(216, self.b)
}
```

The test `case` must named `TestXXX`. case function has only one argument: `*gunit.Case`.

The `SetUp` and `TearDown` function are both optional.

## Value Check

Check equals:
```go
func (self *Case)AssertEquals(lhs, rhs interface{})

func (self *Case)ExpectEquals(lhs, rhs interface{})
```

Check `true` or `false`:
```go
func (self *Case)AssertTrue(rhs bool)
func (self *Case)AssertFalse(rhs bool)

func (self *Case)ExpectTrue(rhs bool)
func (self *Case)ExpectFalse(rhs bool)
```

Check `nil` or NOT `nil`:
```go
func (self *Case)AssertNil(rhs interface{})
func (self *Case)AssertNotNil(rhs interface{})

func (self *Case)ExpectNil(rhs interface{})
func (self *Case)ExpectNotNil(rhs interface{})
```

> `Assert` can break test case, but `Expect` not.

### How to check multi-return function?

Your can use the `Batch` checker:

```go
func (self *MySuite) TestBatchCheck(c *gunit.Case) {
    c.Batch(foo()).
        AssertEquals(1).    // check first value
        Skip(1).            // skip second value
        AssertEquals("ok"). // check third value
        AssertNil()         // and last one
}

func foo() (int, int, string, error) {
    return 1, -1, "ok", nil
}
```

## Panic Test

Expect panic will be happen:
```go
func (self *Case) ExpectPanic(lhs interface{})
```

Assert panic:
```go
func (self *Case) AssertPanic(lhs, fn func())
```

Example:
```go
func (self *MySuite) TestExpectPanic(c *gunit.Case) {
    defer c.ExpectPanic("panic!") // must be in `defer` statement

    panic("panic!")
}

func (self *MySuite) TestScopedPanic(c *gunit.Case) {
    c.AssertPanic("error!", func () {
        c.T.Log("panic incoming!")
        panic("error!")
    })
}
```