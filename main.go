package main

import (
	"flag"
	"fmt"
	i "github.com/traPtitech/traQ-Bench/init"
	r "github.com/traPtitech/traQ-Bench/run"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "init":
		i.Init()
	case "run":
		r.Run()
	case "userdump":
		i.DumpUsers()
	default:
		fmt.Println("Error unknown argument")
	}
}
