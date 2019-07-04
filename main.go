package main

import (
	i "github.com/traPtitech/traQ-Bench/init"
	r "github.com/traPtitech/traQ-Bench/run"
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "init":
			i.Init()
	case "run":
			r.Run()
	default:
		fmt.Println("Error unknown argument")
	}
}
