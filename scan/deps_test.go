package scan

import (
	"os"
	"strings"
	"testing"

	. "launchpad.net/gocheck"

	"github.com/ernestokarim/closurer/config"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

type DepsSuite struct{}

var _ = Suite(&DepsSuite{})
var curDir string

func (s *DepsSuite) SetUpSuite(c *C) {
	var err error
	curDir, err = os.Getwd()
	if err != nil {
		c.Error(err)
		return
	}

	if err := os.Chdir("/home/ernesto/projects/geohistoria"); err != nil {
		c.Error(err)
		return
	}

	config.ConfPath = "client/config.xml"
	config.BuildTargets = "production"

	if err := config.Load(); err != nil {
		c.Error(err)
		return
	}
}

func (s *DepsSuite) TearDownSuite(c *C) {
	if err := os.Chdir(curDir); err != nil {
		c.Error(err)
		return
	}
}

func (s *DepsSuite) BenchmarkGeneration(c *C) {
	conf := config.Current()

	for i := 0; i < c.N; i++ {
		depstree, err := NewDepsTree("compile")
		if err != nil {
			c.Error(err)
			return
		}

		namespaces := []string{}
		for _, input := range conf.Js.Inputs {
			if strings.Contains(input.File, "_test") {
				continue
			}

			ns, err := depstree.GetProvides(input.File)
			if err != nil {
				c.Error(err)
				return
			}
			namespaces = append(namespaces, ns...)
		}

		if _, err := depstree.GetDependencies(namespaces); err != nil {
			c.Error(err)
			return
		}
	}
}
