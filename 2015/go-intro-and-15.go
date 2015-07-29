package main

import "fmt"

func slowOrLongRunningFunction() {
}

func goroutineExample() {
	// "Blocking" function call
	slowOrLongRunningFunction()

	// "Non-Blocking" function call
	go slowOrLongRunningFunction()
}

// Interface Start OMIT
type Response interface {
	Message() string
}

func PrintStruct(s Response) {
	fmt.Printf("%s\n", s.Message())
}

type MyStruct struct {
	msg string
}

func (s MyStruct) Message() string {
	return s.msg
}

func main() {
	s := MyStruct{"Hello World"}
	PrintStruct(s)
}

// Interface End OMIT
