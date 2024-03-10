package main

import (
	"fmt"
	"rsc.io/quote"
	"github.com/mikew686/rpidemo/internal/sanity"
)

func main() {
	fmt.Println("Sanity says:", sanity.Sanity())
    fmt.Println("Quote:", quote.Go())
}
