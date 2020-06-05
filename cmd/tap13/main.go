package main

import (
	"fmt"
	"github.com/mpontillo/tap13"
	"os"
)

func main() {
	args := os.Args[1:]
	for _, arg := range args {
		fmt.Println(arg)
		contents := tap13.ReadFile(arg)
		results := tap13.Parse(contents)
		fmt.Println(results)
	}
}
