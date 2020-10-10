package main

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/swagger"
)

func main() {
	s := g.Server()
	s.Plugin(&swagger.Swagger{})
	s.SetPort(8199)
	s.Run()
}
