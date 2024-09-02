#include <fstream>
#include <iostream>

enum primitive_types {
  int_8,
  int_16,
  int_32,
  int_64,
  uint_8,
  uint_16,
  uint_32,
  uint_64,
  float_32,
  float_64
};

// compound types
enum compound_types { struct_type, array_type };

//    grammar:
//
//    type s struct {
//      a int_8
//    }
//
//    type b = []int_8
//
//    var a int_8
//
//    func makeFoo(T) T {
//      return struct {
//        a T
//      }
//    }
//
//    type fooInt = makeFoo(int)

int main(int argc, char *argv[]) {
  if (argc < 2) {
    std::cerr << "driver filename" << std::endl;
    return -1;
  }
  std::ifstream in(argv[1]);
  std::istream_iterator<std::string> begin(in), end;
  for (; begin != end; ++begin) {
    std::cout << *begin << std::endl;
  }
  return 0;
}
