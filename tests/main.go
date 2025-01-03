package main

type Composite struct {
  a []int
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
  for x:=0; x<10;x++ {
    if len(a) == 0 {
    } else if len(a) == 0 {
    }
  }
  for _, x := range a {
  }
  return 5
}

func bar2() (int16,int16) {
  return 10,20
}

func sink(p int8) {
}

func foo() {
  var a int8
  var b, c int16
  a = 1
  a = a + 5
  d := 10
  a = bar()
  b,c = bar2()
  if a == 1 && b == 10 {
    a = 2
    var aa int8
    aa = bar()
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
  foo()
}
