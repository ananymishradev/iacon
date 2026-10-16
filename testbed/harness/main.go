package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	help := flag.NewFlagSet("harness", flag.ExitOnError)
	help.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: harness [config]")
		help.PrintDefaults()
	}
	if err := help.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if help.NArg() > 0 && help.Arg(0) == "help" {
		help.Usage()
	}
}
