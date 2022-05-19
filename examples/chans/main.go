package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)

	for i := 0; i < 10; i++ {
		select {
		case ch <- i:
		default:
		}
	}

	time.Sleep(time.Second)

	for i := 0; i < 4; i++ {
		fmt.Println(<-ch)

	}
	// close(ch)
	// fmt.Println(<-ch)
	// fmt.Println(<-ch)
	// fmt.Println(<-ch)
	// fmt.Println(<-ch)
}
