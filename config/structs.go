package config

type Config struct {
	Build string `xml:"build,attr"`

	Output  OutputNode  `xml:"output"`
	Js      JsNode      `xml:"js"`
	Gss     GssNode     `xml:"gss"`
	Soy     SoyNode     `xml:"soy"`
	Library LibraryNode `xml:"library"`
}

type OutputNode struct {
	Js  string `xml:"js,attr"`
	Css string `xml:"css,attr"`
}

type JsNode struct {
	Root     string `xml:"root,attr"`
	Compiler string `xml:"compiler,attr"`

	Checks  ChecksNode     `xml:"checks"`
	Targets []JsTargetNode `xml:"target"`
	Inputs  []InputNode    `xml:"input"`
	Externs []ExternNode   `xml:"extern"`
}

func (n *JsNode) CurTarget() *JsTargetNode {
	for _, t := range n.Targets {
		if t.Name == Target {
			return &t
		}
	}
	return nil
}

type ChecksNode struct {
	Errors   []CheckNode `xml:"error"`
	Warnings []CheckNode `xml:"warning"`
	Offs     []CheckNode `xml:"off"`
}

type CheckNode struct {
	Name string `xml:"name,attr"`
}

type JsTargetNode struct {
	Name  string `xml:"name,attr"`
	Mode  string `xml:"mode,attr"`
	Level string `xml:"level,attr"`

	Defines []DefineNode `xml:"define"`
}

type DefineNode struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type InputNode struct {
	File string `xml:"file,attr"`
}

type ExternNode struct {
	File string `xml:"file,attr"`
}

type GssNode struct {
	Root     string `xml:"root,attr"`
	Compiler string `xml:"compiler,attr"`

	Targets []GssTargetNode `xml:"target"`
	Funcs   []FuncNode      `xml:"func"`
}

func (n *GssNode) CurTarget() *GssTargetNode {
	for _, t := range n.Targets {
		if t.Name == Target {
			return &t
		}
	}
	return nil
}

type GssTargetNode struct {
	Name   string `xml:"name,attr"`
	Rename string `xml:"rename,attr"`

	Defines []DefineNode `xml:"define"`
}

type SoyNode struct {
	Root     string `xml:"root,attr"`
	Compiler string `xml:"compiler,attr"`
}

type LibraryNode struct {
	Root string `xml:"root,attr"`
}

type FuncNode struct {
	Name string `xml:"name,attr"`
}
