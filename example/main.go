package main

import (
	"fmt"
	"github.com/ernestokarim/closurer/config"
)

func main() {
	conf, err := config.Load("test.xml")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%+v\n", conf)
}
