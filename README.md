# ULang

This project is a personal experiment aimed at implementing a Go language transpiler that provides a foundation for writing portable libraries.
Currently, only a limited set of primitives is supported for transpilation. The goal is to gradually extend this subset over time.
Ultimately, some features might remain unsupported due to the lack of corresponding primitives on the target platform.

## Usage

To build the project, run the following command:

```bash
  go generate
  go run . --source=[directory] --output=[file]
```

The `--source` flag specifies the directory containing the source files, while the `--output` flag specifies the output file where the transpiled code will be written.
The transpiler will process all `.go` files in the specified source directory and generate the corresponding output file.

## Example
To transpile the source files located in the `src` directory and write the output to `output`, use the following command:

```bash
  go run . --source=./../libs/uql --output=uql
```