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
  d := 10
  a = bar()
  b,c = bar2()
  if a == 1 && b == 10 {
  }
  if b == 10 {
  }
  if c == 20 {
  }
  if d == 10 {
  }
}

func main() {
  foo()
}