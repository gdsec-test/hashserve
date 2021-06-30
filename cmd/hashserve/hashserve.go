package main

import (
	"context"
	"log"

	"github.com/gdcorp-infosec/hashserve/pkg/cmd/hashserve"
)

func main() {
	if err := hashserve.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
