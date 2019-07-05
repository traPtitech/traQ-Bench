package main

import (
	"flag"
	i "github.com/traPtitech/traQ-Bench/init"
	r "github.com/traPtitech/traQ-Bench/run"
	"log"
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	switch flag.Arg(0) {
	case "init":
		i.Init()
	case "run":
		r.Run()
	case "userdump":
		i.DumpUsers()
	default:
		log.Println("Error unknown argument")
	}
}
