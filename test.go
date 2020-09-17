package main

import (
	"github.com/gogf/gf-swagger/swagger"
	"github.com/gogf/gf/frame/g"
)

func main() {
	s := g.Server()
	s.Plugin(&swagger.Swagger{})
	s.SetPort(8199)
	s.Run()
}
