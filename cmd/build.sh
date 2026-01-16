#!/bin/bash

# Build script for ULC with astyle CGO integration

set -e

echo "Generating code..."
go generate ./...

echo "Building astyle static library..."
cd astyle
make -f Makefile.cgo clean
make -f Makefile.cgo
cd ..

echo "Building ULC..."
go build -v .

echo "Build completed successfully!"