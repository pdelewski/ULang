package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"
)

type CPPEmitter struct {
	file *os.File
}

func (*CPPEmitter) lowerToBuiltins(selector string) string {
	switch selector {
	case "fmt.Sprintf":
		return "string_format"
	case "fmt.Println":
		return "println"
	case "fmt.Printf":
		return "printf"
	case "fmt.Print":
		return "printf"
	case "len":
		return "std::size"
	}
	return selector
}

func (e *CPPEmitter) emitToFile(s string) error {
	_, err := e.file.WriteString(s)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}

func (e *CPPEmitter) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
}

func (v *CPPEmitter) SetFile(file *os.File) {
	v.file = file
}

func (cppe *CPPEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	if e.Kind == token.STRING {
		e.Value = strings.Replace(e.Value, "\"", "", -1)
		if e.Value[0] == '`' {
			e.Value = strings.Replace(e.Value, "`", "", -1)
			cppe.emitToFile(cppe.emitAsString(fmt.Sprintf("R\"(%s)\"", e.Value), 0))
		} else {
			cppe.emitToFile(cppe.emitAsString(fmt.Sprintf("\"%s\"", e.Value), 0))
		}
	} else {
		cppe.emitToFile(cppe.emitAsString(e.Value, 0))
	}
}
func (*CPPEmitter) PostVisitBasicLit(node *ast.BasicLit, indent int) {

}

func (cppe *CPPEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	var str string
	name := e.Name
	name = cppe.lowerToBuiltins(name)
	if name == "nil" {
		str = cppe.emitAsString("{}", indent)
		cppe.emitToFile(str)
	} else {
		if n, ok := typesMap[name]; ok {
			str = cppe.emitAsString(n, indent)
			cppe.emitToFile(str)
		} else {
			str = cppe.emitAsString(name, indent)
			cppe.emitToFile(str)
		}
	}
}

func (*CPPEmitter) PostVisitIdent(node *ast.Ident, indent int) {}
