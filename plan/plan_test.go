package plan

import (
	"testing"

	yaml "launchpad.net/goyaml"
)

func TestPlanSanity(t *testing.T) {

	r := &Relation{
		HostName:   "h1",
		DBName:     "d1",
		SchemaName: "t1",
	}

	node := Alias{
		nodeBase: nodeBase{Children: []Node{r}},
	}

	if a, ok := node.nodeBase.Children[0].(*Relation); a != r || !ok {
		t.Fail()
	}

	if node.Limit != nil || node.Filter != nil {
		t.Fail()
	}
}

func TestPlanVisit(t *testing.T) {
	v := &testVisitor{t}

	r := &Relation{
		HostName:   "h2",
		DBName:     "d2",
		SchemaName: "t2",
	}

	node := Alias{
		nodeBase: nodeBase{Children: []Node{r}},
	}

	f := true
	node.Accept(v, &f)

	if s, err := yaml.Marshal(&node); err != nil {
		t.Fail()
	} else {
		t.Log(string(s))
	}
}

type testVisitor struct {
	t *testing.T
}

func (self *testVisitor) OnMerge(node *Merge, f *bool) {
	self.t.Log(node)
}
func (self *testVisitor) OnProject(node *Project, f *bool) {
	self.t.Log(node)
}
func (self *testVisitor) OnSelect(node *Select, f *bool) {
	self.t.Log(node)
}
func (self *testVisitor) OnRelation(node *Relation, f *bool) {
	self.t.Log(node)
}
func (self *testVisitor) OnAlias(node *Alias, f *bool) {
	self.t.Log(node)
}
