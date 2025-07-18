package main

import (
	"fmt"

	"github.com/alitto/pond/v2"
)

func main() {
	pool := pond.NewPool(100)

	for i := 0; i < 1000; i++ {
		i := i
		pool.Submit(func() {
			fmt.Printf("Running task #%d\n", i)
		})
	}

	pool.StopAndWait()
}
