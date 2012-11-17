package main

import (
	"os"
	"testing"

	. "launchpad.net/gocheck"

	"github.com/ernestokarim/closurer/config"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

type ServeSuite struct{}

var _ = Suite(&ServeSuite{})
var curDir string

func (s *ServeSuite) SetUpSuite(c *C) {
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

	if err := config.Load(); err != nil {
		c.Error(err)
		return
	}
}

func (s *ServeSuite) TearDownSuite(c *C) {
	if err := os.Chdir(curDir); err != nil {
		c.Error(err)
		return
	}
}

// This benchmark it's intended to run in a production-like environmente,
// and save the cpu/mem profile of the real server.
func (s *ServeSuite) BenchmarkGeneration(c *C) {
	serve()
}
