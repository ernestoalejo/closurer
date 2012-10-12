package main

import (
	"github.com/ernestokarim/closurer/app"
)

func Home(r *app.Request) error {
	return r.ExecuteTemplate([]string{"home"}, nil)
}
