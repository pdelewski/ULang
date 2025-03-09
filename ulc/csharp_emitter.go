package main

import (
	"os"
)

type CSharpEmitter struct {
	file *os.File
	Emitter
	insideForPostCond bool
	assignmentToken   string
}

func (*CSharpEmitter) lowerToBuiltins(selector string) string {
	return ""
}

func (e *CSharpEmitter) emitToFile(s string) error {
	return nil
}

func (e *CSharpEmitter) emitAsString(s string, indent int) string {
	return ""
}

func (cppe *CSharpEmitter) SetFile(file *os.File) {
	cppe.file = file
}

func (cppe *CSharpEmitter) GetFile() *os.File {
	return cppe.file
}
