package main

import (
	"fmt"
)

func HomeHandler(r *Request) error {
	fmt.Fprintln(r.W, "Hello World!")
	return nil
}
