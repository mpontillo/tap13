package main

import (
	"fmt"
	"os"

	"github.com/mpontillo/tap13"
	util "github.com/mpontillo/tap13/internal"
)

func main() {
	args := os.Args[1:]
	for _, arg := range args {
		fmt.Println(arg)
		contents := util.ReadFile(arg)
		results := tap13.Parse(contents)
		fmt.Println(results)
	}
}
