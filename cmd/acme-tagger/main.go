package main

import (
	"log"

	"github.com/jcowgar/acme-utils/internal/tagger"
)

func main() {
	c, e := tagger.LoadConfiguration()
	if e != nil {
		log.Printf("failed loading configuration: %v\n", e)
		return
	}

	tagger.ExecuteLogLoop(c)
}
