package main

import "fmt"

type Composite struct {
	a []int
}

func testBasicConstructs() int8 {
	testSliceOperations()
	testLoopConstructs()
	testBooleanLogic()
	return 5
}

func testFunctionCalls() (int16, int16) {
	return testFunctionVariables()
}

func testSliceOperations() {
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
}

func testLoopConstructs() {
	var a []int

	for x := 0; x < 10; x++ {
		if !(len(a) == 0) {
		} else if len(a) == 0 {
		}
	}

	for _, x := range a {
		if x == 0 {
		}
	}
}

func testBooleanLogic() {
	b := false
	if !b {
	}
}

func testFunctionVariables() (int16, int16) {
	x := []func(int, int){
		func(a int, b int) {
			fmt.Println(a)
			fmt.Println(b)
		},
	}

	f := x[0]
	f(10, 20)
	x[0](20, 30)

	if len(x) == 0 {
	}

	return 10, 20
}

func sink(p int8) {
}

func testArrayInitialization() {
	a := []int8{}
	if len(a) == 0 {
	}

	b := []int8{1, 2, 3}
	if len(b) == 0 {
	}
}

func testSliceExpressions() {
	a := []int8{1, 2, 3}
	b := a[1:]
	if len(b) == 0 {
	}
}

func testCompleteLanguageFeatures() {
	var a int8
	var b, c int16

	a = 1
	a = a + 5
	d := 10

	a = testBasicConstructs()
	b, c = testFunctionCalls()

	if (a == 1) && (b == 10) {
		a = 2
		var aa int8
		aa = testBasicConstructs()
		sink(aa)

		if a == 5 {
			a = 10
		}
	} else {
		a = 3
	}

	if b == 10 {
	}
	if c == 20 {
	}
	if d == 10 {
	}
}

func main() {
	fmt.Print("Hello\n")
	testCompleteLanguageFeatures()
	testArrayInitialization()
	testSliceExpressions()
}
