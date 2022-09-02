package main

import (
	"fmt"
	"os"
	"os/signal"
)

func pause() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	s := <-c
	fmt.Printf("Shutting down, got signal: %s", s)
}
