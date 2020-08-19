package main

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/ysh86/svg"
)

func printIdent(level int) {
	for ; level > 0; level-- {
		fmt.Print(" ")
	}
}

func main() {
	dec := xml.NewDecoder(os.Stdin)

	svg := new(svg.Root)
	err := svg.Parse(dec)
	if err != nil {
		panic(err)
	}

	fmt.Println(svg)
	for _, g := range svg.Groups {
		printIdent(1)
		fmt.Println(g)
		for _, gg := range g.Groups {
			printIdent(2)
			fmt.Println(gg)
			for _, p := range gg.Paths {
				printIdent(3)
				fmt.Println(p)
				for _, c := range p.D {
					printIdent(4)
					fmt.Println(c)
				}
			}
		}
	}
	fmt.Println("done")
}
