package scan

import (
	"github.com/ernestokarim/closurer/domain"
)

// Return true if s is in lst.
func In(lst []string, s string) bool {
	for _, v := range lst {
		if v == s {
			return true
		}
	}
	return false
}

// Return true if s is in lst.
func InSource(lst []*domain.Source, s *domain.Source) bool {
	for _, v := range lst {
		if v == s {
			return true
		}
	}
	return false
}

// Return true if name is a directory that must be scanned recursively
// for any type of interesting files.
func IsValidDir(name string) bool {
	return name != ".svn" && name != ".hg" && name != ".git"
}
