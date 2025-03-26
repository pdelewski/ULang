package main

import "fmt"

func main() {
	var a []int

	b := len(a)

	fmt.Println(b)

	c := []int{1, 2, 3}

	d := len(c)

	fmt.Println(d)
}
