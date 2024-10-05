# ULang

This project is a personal experiment aimed at implementing a language with a very limited set of primitives, 
yet full expressiveness, that can be easily mapped to other programming languages, providing a foundation for writing portable libraries.
The language currently is a subset of golang which means that it is implemented using golang syntax and its abstractions.

## Primitvies types

* int8
* int16
* int32
* int64
* uint8
* uint16
* uint32
* uint64
* float32
* float64

## Compound types

* array
* struct

Array use COW (copy on write) semantics.

## Control flow statements

* if
* for

## Functions

All parameters are passed by value except arrays. Changing struct instance means returning new one,
however that might be implemented using pointers in backend programming language that have them.

```golang
type A struct {
  x int
}

func process(a A) A {
  a := A{x:2}
  return a;
}
```

## Allowed operations (semantically correct ULang subset)

* primitives

```
  x := 1 // x inferred as int8
  x = 2  // mutating x
```

* structs

```
  a := A{x:1}
  a.x = 2
```

* arrays
