package main

import (
	"fmt"
	"go/ast"
	"os"
	"strings"
	"unicode"
)

type CSharpEmitter struct {
	file *os.File
	Emitter
	insideForPostCond bool
	assignmentToken   string
	forwardDecls      bool
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s // Return empty string if input is empty
	}

	// Convert string to rune slice to handle Unicode characters
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0]) // Change the first character to uppercase

	return string(runes) // Convert runes back to string
}

func (*CSharpEmitter) lowerToBuiltins(selector string) string {
	return ""
}

func (e *CSharpEmitter) emitToFile(s string) error {
	_, err := e.file.WriteString(s)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

func (e *CSharpEmitter) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
}
func (cppe *CSharpEmitter) SetFile(file *os.File) {
	cppe.file = file
}

func (cppe *CSharpEmitter) GetFile() *os.File {
	return cppe.file
}

func (cppe *CSharpEmitter) PreVisitProgram(indent int) {
	outputFile := "./output.cs"
	var err error
	cppe.file, err = os.Create(outputFile)
	cppe.SetFile(cppe.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	_, err = cppe.file.WriteString("using System;\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	cppe.insideForPostCond = false
}

func (cppe *CSharpEmitter) PostVisitProgram(indent int) {
	cppe.file.Close()
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}
	if node.Name.Name == "main" {
		str := cppe.emitAsString(fmt.Sprintf("class %s\n", capitalizeFirst(node.Name.Name)), 0)
		err := cppe.emitToFile(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
		err = cppe.emitToFile("{\n\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclSignatures(indent int) {
	cppe.forwardDecls = true
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignatures(indent int) {
	cppe.forwardDecls = false
}

func (cppe *CSharpEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	if cppe.forwardDecls {
		return
	}
	if node.Name.Name == "main" {
		str := cppe.emitAsString(fmt.Sprintf("\n} // class %s\n\n", capitalizeFirst(node.Name.Name)), 0)
		err := cppe.emitToFile(str)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}
}

func (cppe *CSharpEmitter) PreVisitFuncDeclName(node *ast.Ident, indent int) {
	if cppe.forwardDecls {
		return
	}
	if node.Name == "main" {
		str := cppe.emitAsString(fmt.Sprintf("public static void %s()\n", capitalizeFirst(node.Name)), indent)
		cppe.emitToFile(str)
	}
}
