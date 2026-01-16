CXX = g++
CXXFLAGS = -std=c++17 -I/Users/pdelewski/Projects/ULang/runtime $(shell sdl2-config --cflags)
LDFLAGS = $(shell sdl2-config --libs)

TARGET = test-check
SRCS = test-check.cpp

all: $(TARGET)

$(TARGET): $(SRCS)
	$(CXX) $(CXXFLAGS) -o $@ $^ $(LDFLAGS)

clean:
	rm -f $(TARGET)

.PHONY: all clean
