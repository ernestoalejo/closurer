package main

import ()

func In(lst []string, s string) bool {
	for _, v := range lst {
		if v == s {
			return true
		}
	}
	return false
}

func InSource(lst []*Source, s *Source) bool {
	for _, v := range lst {
		if v == s {
			return true
		}
	}
	return false
}
