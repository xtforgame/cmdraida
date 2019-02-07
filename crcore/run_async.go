package crcore

import (
// "fmt"
)

// https://medium.com/@nikolay.bystritskiy/how-i-tried-to-do-things-asynchronously-in-golang-40e0c1a06a66

// func CancelableAsync(task func() interface{}, cancel chan struct{}) chan interface{} {
// 	// out := make(chan interface{})
// 	out := make(chan interface{}, 1)
// 	go func() {
// 		defer close(out)
// 		select {
// 		// Received a signal to abandon further processing
// 		case <-cancel:
// 			return
// 		// Got some result
// 		case out <- task():
// 		}
// 	}()
// 	return out
// }

// func CancelableAsync(task func() interface{}, cancel chan struct{}) {
// 	// out := make(chan interface{})
// 	out := make(chan interface{}, 1)
// 	go func() {
// 		defer close(out)
// 		select {
// 		// Received a signal to abandon further processing
// 		case <-cancel:
// 			return
// 		// Got some result
// 		case out <- task():
// 		}
// 	}()

// 	go func() {
// 		result, ok := (<-out).(*TaskBase)
// 		fmt.Println("result, ok :", result, ok)
// 	}()
// }

func CancelableAsync(task func() interface{}, callback func(interface{}), cancel chan interface{}) {
	// out := make(chan interface{})
	out1 := make(chan interface{}, 1)
	out2 := make(chan interface{}, 1)
	go func() {
		out1 <- task()
	}()
	go func() {
		select {
		// Received a signal to abandon further processing
		case c := <-cancel:
			out2 <- c
			break
		// Got some result
		case r := <-out1:
			out2 <- r
			break
		}
		callback(<-out2)
	}()
}
