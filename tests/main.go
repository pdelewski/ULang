package main

func bar() int8 {
  return 5
}


func foo() {
  var a int8
  var b, c int16
  a = 1
  a = bar()
}

func main() {
  foo()
}