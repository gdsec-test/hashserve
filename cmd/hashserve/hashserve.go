package main

import (
	"context"
	"github.com/gdcorp-infosec/hashserve/pkg/cmd/hashserve"
	"log"
)

func main() {
	if err := hashserve.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
