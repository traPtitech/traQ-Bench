package main

import (
	"flag"
	constant "github.com/traPtitech/traQ-Bench/const"
	i "github.com/traPtitech/traQ-Bench/init"
	r "github.com/traPtitech/traQ-Bench/run"
	"log"
	"strconv"
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	switch flag.Arg(0) {
	case "init":
		i.Init()
	case "run":
		maxStr := flag.Arg(1)
		if max, err := strconv.Atoi(maxStr); err == nil {
			r.Run(max)
		} else {
			r.Run(constant.MaxUsers)
		}
	case "userdump":
		i.DumpUsers()
	default:
		log.Println("Error unknown argument")
	}
}
