package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var p string
	flag.StringVar(&p, "p", "", "<pkg path>:<ident>[:[<field>|<method>()]]")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `ged [options] [packages]

Options:`)
		flag.PrintDefaults()
	}
	flag.Parse()
	pattern, err := parsePattern(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse pattern: %v", err)
		os.Exit(1)
	}
	matches, err := pattern.FindDepInPackages(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println(matches)
}
