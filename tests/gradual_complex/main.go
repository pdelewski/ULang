package main

import "fmt"

type Composite struct {
        a []int
}

func sink(p int8) {
}

func bar() int8 {
        var a []int
        c := Composite{}
        if len(a) == 0 {
        } else {
                if a[0] == 0 {
                        a[0] = 1
                }
        }
        if len(c.a) == 0 {
        }
        for x := 0; x < 10; x++ {
                if !(len(a) == 0) {
                } else if len(a) == 0 {
                }
        }
        for _, x := range a {
                if x == 0 {
                }
        }
        b := false
        if !b {
        }
        return 5
}

func arraytype() {
	a := []int8{}
	if len(a) == 0 {
	}
	b := []int8{1, 2, 3}
	if len(b) == 0 {
	}
}


func main() {
	fmt.Print("Hello\n")
}
