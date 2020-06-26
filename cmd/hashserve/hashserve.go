package main

import (
	"context"
	"github.secureserver.net/digital-crimes/hashserve/pkg/cmd/hashserve"
	"log"
)

func main() {
	if err := hashserve.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}