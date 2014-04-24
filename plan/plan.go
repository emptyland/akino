package plan

type Merge struct {
	nodeBase
}

func (self *Merge) Children() []Node {
	return self.nodeBase.Children
}

func (self *Merge) Accept(v Visitor, f *bool) {
	v.OnMerge(self, f)
	visitChildrenIfNeed(self, v, f)
}

type Project struct {
	nodeBase
	Columns []Node
}

func (self *Project) Children() []Node {
	return self.nodeBase.Children
}

func (self *Project) Accept(v Visitor, f *bool) {
	v.OnProject(self, f)
	visitChildrenIfNeed(self, v, f)
}

type Select struct {
	nodeBase
}

func (self *Select) Children() []Node {
	return self.nodeBase.Children
}

func (self *Select) Accept(v Visitor, f *bool) {
	v.OnSelect(self, f)
	visitChildrenIfNeed(self, v, f)
}

type Relation struct {
	nodeBase `yaml:"Relation"`
	// ... metadata
	HostName   string `yaml:"hostName"`
	DBName     string `yaml:"dbName"`
	SchemaName string `yaml:"schemaName"`
}

func (self *Relation) Children() []Node {
	return self.nodeBase.Children
}

func (self *Relation) Accept(v Visitor, f *bool) {
	v.OnRelation(self, f)
	visitChildrenIfNeed(self, v, f)
}

type Alias struct {
	nodeBase `yaml:"Alias"`
	Name     string
}

func (self *Alias) Children() []Node {
	return self.nodeBase.Children
}

func (self *Alias) Accept(v Visitor, f *bool) {
	v.OnAlias(self, f)
	visitChildrenIfNeed(self, v, f)
}

type nodeBase struct {
	Limit    *Limit
	Filter   *Filter
	Children []Node
}

type Node interface {
	Children() []Node
	Accept(v Visitor, f *bool)
}

type Limit struct {
	Offset int64
	Limit  int64
}

type Filter struct {
}

type Visitor interface {
	OnMerge(node *Merge, f *bool)
	OnProject(node *Project, f *bool)
	OnSelect(node *Select, f *bool)
	OnRelation(node *Relation, f *bool)
	OnAlias(node *Alias, f *bool)
}

func visitChildrenIfNeed(node Node, visitor Visitor, f *bool) {
	if *f && node.Children() != nil {
		for _, child := range node.Children() {
			child.Accept(visitor, f)
			if !*f {
				return
			}
		}
	}
}
