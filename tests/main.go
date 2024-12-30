package main

func bar() int8 {
  return 5
}

func bar2() (int16,int16) {
  return 10,20
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