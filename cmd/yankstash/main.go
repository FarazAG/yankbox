package main

import (
	"fmt"
	"log"

	"github.com/FarazAG/yankstash/internal/clipboard"
)

func main() {
	items, err := clipboard.History()
	if err != nil {
		log.Fatal(err)
	}

	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item.Text)
	}
}
