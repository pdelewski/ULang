package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

var destTypes = []string{"sbyte", "short", "int", "long", "byte", "ushort", "object", "string"}

var csTypesMap = map[string]string{
	"int8":   destTypes[0],
	"int16":  destTypes[1],
	"int32":  destTypes[2],
	"int64":  destTypes[3],
	"uint8":  destTypes[4],
	"uint16": destTypes[5],
	"any":    destTypes[6],
	"string": destTypes[7],
}

type AliasRepr struct {
	PackageName string // Package name of the alias
	TypeName    string
}

type Alias struct {
	PackageName    string
	representation []AliasRepr // Representation of the alias
	UnderlyingType string      // Underlying type of the alias as string for now  It's type to what the alias points to
}

type CSharpEmitter struct {
	Output string
	file   *os.File
	Emitter
	pkg               *packages.Package
	insideForPostCond bool
	assignmentToken   string
	forwardDecls      bool
	shouldGenerate    bool
	numFuncResults    int
	aliases           map[string]Alias
	currentPackage    string
	buffer            bool
	isArray           bool
	arrayType         string
	isTuple           bool
	gir               GoFIR
}

func (*CSharpEmitter) lowerToBuiltins(selector string) string {
	switch selector {
	case "fmt":
		return ""
	case "Sprintf":
		return "Formatter.Sprintf"
	case "Println":
		return "Console.WriteLine"
	case "Printf":
		return "Formatter.Printf"
	case "Print":
		return "Formatter.Printf"
	case "len":
		return "SliceBuiltins.Length"
	case "append":
		return "SliceBuiltins.Append"
	}
	return selector
}

func (cse *CSharpEmitter) emitAsString(s string, indent int) string {
	return strings.Repeat(" ", indent) + s
}
func (cse *CSharpEmitter) SetFile(file *os.File) {
	cse.file = file
}

func (cse *CSharpEmitter) GetFile() *os.File {
	return cse.file
}

func (cse *CSharpEmitter) PreVisitProgram(indent int) {
	cse.gir.pointerAndPositionVec = make([]PointerAndPosition, 0)
	cse.aliases = make(map[string]Alias)
	outputFile := cse.Output
	cse.shouldGenerate = true
	var err error
	cse.file, err = os.Create(outputFile)
	cse.SetFile(cse.file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	_, err = cse.file.WriteString("using System;\nusing System.Collections;\nusing System.Collections.Generic;\n\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	builtin := `public static class SliceBuiltins
{
  public static List<T> Append<T>(this List<T> list, T element)
  {
    var result = list != null ? new List<T>(list) : new List<T>();
    result.Add(element);
    return result;
  }

  public static List<T> Append<T>(this List<T> list, params T[] elements)
  {
    var result = list != null ? new List<T>(list) : new List<T>();
    result.AddRange(elements);
    return result;
  }

  public static List<T> Append<T>(this List<T> list, List<T> elements)
  {
    var result = list != null ? new List<T>(list) : new List<T>();
    result.AddRange(elements);
    return result;
  }

  // Fix: Ensure Length works for collections and not generic T
  public static int Length<T>(ICollection<T> collection)
  {
    return collection == null ? 0 : collection.Count;
  }
  public static int Length(string s)
  {
    return s == null ? 0 : s.Length;
  }
}
public class Formatter {
    public static void Printf(string format, params object[] args)
    {
        int argIndex = 0;
        string converted = "";
        List<object> formattedArgs = new List<object>();

        for (int i = 0; i < format.Length; i++)
        {
            if (format[i] == '%' && i + 1 < format.Length)
            {
                char next = format[i + 1];
                switch (next)
                {
                    case 'd':
                    case 's':
                    case 'f':
                        converted += "{" + argIndex + "}";
                        formattedArgs.Add(args[argIndex]);
                        argIndex++;
                        i++; // skip format char
                        continue;
                    case 'c':
                        converted += "{" + argIndex + "}";
                        object arg = args[argIndex];
                        if (arg is sbyte sb)
                            formattedArgs.Add((char)sb); // sbyte to char
                        else if (arg is int iVal)
                            formattedArgs.Add((char)iVal);
                        else if (arg is char cVal)
                            formattedArgs.Add(cVal);
                        else
                            throw new ArgumentException($"Argument {argIndex} for %c must be a char, int, or sbyte");
                        argIndex++;
                        i++; // skip format char
                        continue;
                }
            }

            converted += format[i];
        }

        converted = converted
            .Replace(@"\n", "\n")
            .Replace(@"\t", "\t")
            .Replace(@"\\", "\\");

        Console.Write(string.Format(converted, formattedArgs.ToArray()));
    }

    public static string Sprintf(string format, params object[] args)
     {
        int argIndex = 0;
        string converted = "";
        List<object> formattedArgs = new List<object>();

        for (int i = 0; i < format.Length; i++)
        {
            if (format[i] == '%' && i + 1 < format.Length)
            {
                char next = format[i + 1];
                switch (next)
                {
                    case 'd':
                    case 's':
                    case 'f':
                        converted += "{" + argIndex + "}";
                        formattedArgs.Add(args[argIndex]);
                        argIndex++;
                        i++; // skip format char
                        continue;
                    case 'c':
                        converted += "{" + argIndex + "}";
                        object arg = args[argIndex];
                        if (arg is sbyte sb)
                            formattedArgs.Add((char)sb); // sbyte to char
                        else if (arg is int iVal)
                            formattedArgs.Add((char)iVal);
                        else if (arg is char cVal)
                            formattedArgs.Add(cVal);
                        else
                            throw new ArgumentException($"Argument {argIndex} for %c must be a char, int, or sbyte");
                        argIndex++;
                        i++; // skip format char
                        continue;
                }
            }

            converted += format[i];
        }
        converted = converted
            .Replace(@"\n", "\n")
            .Replace(@"\t", "\t")
            .Replace(@"\\", "\\");

        return string.Format(converted, formattedArgs.ToArray());
    }
}

`
	str := cse.emitAsString(builtin, indent)
	_ = cse.gir.emitToFileBuffer(str, "")

	cse.insideForPostCond = false
}

func (cse *CSharpEmitter) PostVisitProgram(indent int) {
	emitToFile(cse.file, cse.gir.fileBuffer)
	cse.file.Close()
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatures(indent int) {
	cse.forwardDecls = true
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignatures(indent int) {
	cse.forwardDecls = false
}

func (cse *CSharpEmitter) PreVisitFuncDeclName(node *ast.Ident, indent int) {
	if cse.forwardDecls {
		return
	}
	var str string
	if node.Name == "main" {
		str = cse.emitAsString(fmt.Sprintf("Main"), 0)
	} else {
		str = cse.emitAsString(fmt.Sprintf("%s", node.Name), 0)
	}
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cse.emitAsString("{\n", 1)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	str := cse.emitAsString("}", 1)
	cse.gir.emitToFileBuffer(str, "")
	cse.isArray = false
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if cse.forwardDecls {
		return
	}
	cse.shouldGenerate = true
	str := cse.emitAsString("(", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	if cse.forwardDecls {
		return
	}
	cse.shouldGenerate = false
	str := cse.emitAsString(")", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	if !cse.shouldGenerate {
		return
	}
	var str string
	name := e.Name
	name = cse.lowerToBuiltins(name)
	if name == "nil" {
		str = cse.emitAsString("default", indent)
	} else {
		if n, ok := csTypesMap[name]; ok {
			str = cse.emitAsString(n, indent)
		} else {
			str = cse.emitAsString(name, indent)
		}
	}

	if cse.buffer {
		cse.gir.stack = append(cse.gir.stack, str)
	} else {
		cse.gir.emitToFileBuffer(str, "")
	}
}

func (cse *CSharpEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	str := cse.emitAsString("(", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	str := cse.emitAsString(")", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	cse.gir.stack = append(cse.gir.stack, "@@PreVisitBasicLit")
	var str string
	if e.Kind == token.STRING {
		e.Value = strings.Replace(e.Value, "\"", "", -1)
		if e.Value[0] == '`' {
			e.Value = strings.Replace(e.Value, "`", "", -1)
			str = (cse.emitAsString(fmt.Sprintf("@\"(%s)\"", e.Value), 0))
		} else {
			str = (cse.emitAsString(fmt.Sprintf("@\"%s\"", e.Value), 0))
		}
	} else {
		str = (cse.emitAsString(e.Value, 0))
	}
	cse.gir.stack = append(cse.gir.stack, str)
	cse.buffer = true
}

func (cse *CSharpEmitter) PostVisitBasicLit(e *ast.BasicLit, indent int) {
	cse.gir.stack = mergeStackElements("@@PreVisitBasicLit", cse.gir.stack)
	if len(cse.gir.stack) == 1 {
		cse.gir.emitToFileBuffer(cse.gir.stack[len(cse.gir.stack)-1], "")
		cse.gir.stack = cse.gir.stack[:len(cse.gir.stack)-1]
	}

	cse.buffer = false
}

func (cse *CSharpEmitter) PreVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	cse.gir.emitToFileBuffer("", "@PreVisitDeclStmtValueSpecType")
}

func (cse *CSharpEmitter) PostVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	pointerAndPosition := SearchPointerReverse("@PreVisitDeclStmtValueSpecType", cse.gir.pointerAndPositionVec)
	if pointerAndPosition != nil {
		for aliasName, alias := range cse.aliases {
			if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
				cse.gir.fileBuffer, _ = RewriteFileBufferBetween(cse.gir.fileBuffer, pointerAndPosition.Position, len(cse.gir.fileBuffer), aliasName)
			}
		}
	}
}

func (cse *CSharpEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	str := cse.emitAsString(" ", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	var str string
	if cse.isArray {
		str += " = new "
		str += strings.TrimSpace(cse.arrayType)
		str += "();"
		cse.isArray = false
	} else {
		str += " = default;"
	}
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitGenStructFieldType(node ast.Expr, indent int) {
	str := cse.emitAsString("public", indent+2)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	cse.gir.emitToFileBuffer(" ", "")
	// clean array marker as we should generate
	// initializer only for expression statements
	// not for struct fields
	cse.isArray = false
}

func (cse *CSharpEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	cse.gir.emitToFileBuffer(";\n", "")
}

func (cse *CSharpEmitter) PreVisitPackage(pkg *packages.Package, indent int) {
	name := pkg.Name
	cse.pkg = pkg
	var packageName string
	if name == "main" {
		packageName = "MainClass"
	} else {
		//packageName = capitalizeFirst(name)
		packageName = name
	}
	str := cse.emitAsString(fmt.Sprintf("namespace %s {\n\n", packageName), indent)
	err := cse.gir.emitToFileBuffer(str, "")
	err = cse.gir.emitToFileBuffer("", pkg.Name)
	cse.currentPackage = packageName
	str = cse.emitAsString(fmt.Sprintf("public struct %s {\n\n", "Api"), indent+2)
	err = cse.gir.emitToFileBuffer(str, "")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func RebuildNestedType(reprs []AliasRepr) string {
	if len(reprs) == 0 {
		return ""
	}

	// Start from the innermost type
	result := formatAlias(reprs[len(reprs)-1])
	for i := len(reprs) - 2; i >= 0; i-- {
		result = fmt.Sprintf("%s<%s>", formatAlias(reprs[i]), result)
	}
	return result
}

func formatAlias(r AliasRepr) string {
	if r.PackageName != "" {
		return r.PackageName + "." + r.TypeName
	}
	return r.TypeName
}

func (cse *CSharpEmitter) PostVisitPackage(pkg *packages.Package, indent int) {
	pointerAndPosition := SearchPointerReverse(pkg.Name, cse.gir.pointerAndPositionVec)
	if pointerAndPosition != nil {
		var newStr string
		for aliasKey, aliasVal := range cse.aliases {
			aliasRepr := RebuildNestedType(aliasVal.representation)
			newStr += "using " + aliasKey + " = " + aliasRepr + ";\n"
		}
		newStr += "\n"
		cse.gir.fileBuffer, _ = RewriteFileBuffer(cse.gir.fileBuffer, pointerAndPosition.Position, "", newStr)
	}

	str := cse.emitAsString("}\n", indent+2)
	cse.gir.emitToFileBuffer(str, "")
	err := cse.gir.emitToFileBuffer("}\n", "")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	cse.isArray = false
}

func (cse *CSharpEmitter) PostVisitBlockStmtList(node ast.Stmt, index int, indent int) {
	str := cse.emitAsString("\n", indent)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	if cse.forwardDecls {
		return
	}
	str := cse.emitAsString("\n\n", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitGenStructInfo(node GenTypeInfo, indent int) {
	str := cse.emitAsString(fmt.Sprintf("public struct %s\n", node.Name), indent+2)
	str += cse.emitAsString("{\n", indent+2)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitGenStructInfo(node GenTypeInfo, indent int) {
	str := cse.emitAsString("};\n\n", indent+2)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	if cse.forwardDecls {
		return
	}
	cse.gir.stack = append(cse.gir.stack, "@@PreVisitArrayType")
	str := cse.emitAsString("List", indent)
	cse.gir.stack = append(cse.gir.stack, str)
	str = cse.emitAsString("<", 0)
	cse.gir.stack = append(cse.gir.stack, str)

	cse.buffer = true
}
func (cse *CSharpEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	if cse.forwardDecls {
		return
	}

	cse.gir.stack = append(cse.gir.stack, cse.emitAsString(">", 0))

	cse.gir.stack = mergeStackElements("@@PreVisitArrayType", cse.gir.stack)
	if len(cse.gir.stack) == 1 {
		cse.isArray = true
		cse.arrayType = cse.gir.stack[len(cse.gir.stack)-1]
		cse.gir.emitToFileBuffer(cse.gir.stack[len(cse.gir.stack)-1], "")
		cse.gir.stack = cse.gir.stack[:len(cse.gir.stack)-1]
	}

	cse.buffer = false
}

func (cse *CSharpEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	if cse.forwardDecls {
		return
	}

	cse.buffer = true
	cse.gir.stack = append(cse.gir.stack, "@@PreVisitFuncType")
	var str string
	if node.Results != nil {
		str = cse.emitAsString("Func<", indent)
	} else {
		str = cse.emitAsString("Action<", indent)
	}
	cse.gir.stack = append(cse.gir.stack, str)
}
func (cse *CSharpEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	if cse.forwardDecls {
		return
	}

	// move return type to the end of the stack
	// return type is traversed first therefore it has to be moved
	// to the end of the stack due to C# syntax
	if len(cse.gir.stack) > 2 && cse.numFuncResults > 0 {
		returnType := cse.gir.stack[2]
		cse.gir.stack = append(cse.gir.stack[:2], cse.gir.stack[3:]...)
		cse.gir.stack = append(cse.gir.stack, ",")
		cse.gir.stack = append(cse.gir.stack, returnType)
	}
	cse.gir.stack = append(cse.gir.stack, cse.emitAsString(">", 0))

	cse.gir.stack = mergeStackElements("@@PreVisitFuncType", cse.gir.stack)

	if len(cse.gir.stack) == 1 {
		cse.gir.emitToFileBuffer(cse.gir.stack[len(cse.gir.stack)-1], "")
		cse.gir.stack = cse.gir.stack[:len(cse.gir.stack)-1]
	}
	cse.buffer = false
}

func (cse *CSharpEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.gir.stack = append(cse.gir.stack, str)
	}
}

func (cse *CSharpEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
	if cse.forwardDecls {
		return
	}
	var str string
	scopeOperator := "."
	if ident, ok := node.(*ast.Ident); ok {
		if cse.lowerToBuiltins(ident.Name) == "" {
			return
		}
		// if the identifier is a package name, we need to append "Api." to the scope operator
		obj := cse.pkg.TypesInfo.Uses[ident]
		if obj != nil {
			if _, ok := obj.(*types.PkgName); ok {
				scopeOperator += "Api."
			}
		}
	}

	str = cse.emitAsString(scopeOperator, 0)
	if cse.buffer {
		cse.gir.stack = append(cse.gir.stack, str)
	} else {
		cse.gir.emitToFileBuffer(str, "")
	}

}

func (cse *CSharpEmitter) PreVisitFuncTypeResults(node *ast.FieldList, indent int) {
	if node != nil {
		cse.numFuncResults = len(node.List)
	}
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	if cse.forwardDecls {
		return
	}
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	if cse.forwardDecls {
		return
	}
	cse.gir.emitToFileBuffer(" ", "")
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if cse.forwardDecls {
		return
	}
	if index > 0 {
		str := cse.emitAsString(",", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
	cse.gir.emitToFileBuffer("", "@PreVisitFuncDeclSignatureTypeResultsList")
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	if cse.forwardDecls {
		return
	}
	pointerAndPosition := SearchPointerReverse("@PreVisitFuncDeclSignatureTypeResultsList", cse.gir.pointerAndPositionVec)
	if pointerAndPosition != nil {
		for aliasName, alias := range cse.aliases {
			if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
				cse.gir.fileBuffer, _ = RewriteFileBufferBetween(cse.gir.fileBuffer, pointerAndPosition.Position, len(cse.gir.fileBuffer), aliasName)
			}
		}
	}
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if cse.forwardDecls {
		return
	}

	str := cse.emitAsString("public static ", indent+2)
	cse.gir.emitToFileBuffer(str, "")
	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cse.emitAsString("(", 0)
			cse.gir.emitToFileBuffer(str, "")
		}
	} else {
		str := cse.emitAsString("void", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	if cse.forwardDecls {
		return
	}

	if node.Type.Results != nil {
		if len(node.Type.Results.List) > 1 {
			str := cse.emitAsString(")", 0)
			cse.gir.emitToFileBuffer(str, "")
		}
	}

	str := cse.emitAsString("", 1)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	cse.gir.stack = append(cse.gir.stack, "@@PreVisitTypeAliasName")
	cse.gir.stack = append(cse.gir.stack, cse.emitAsString("using ", indent+2))
	cse.buffer = true
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	cse.buffer = true
	cse.gir.stack = append(cse.gir.stack, " = ")
}

func ParseNestedTypes(s string) []string {
	var result []string
	s = strings.TrimSpace(s)

	for strings.HasPrefix(s, "List<") {
		result = append(result, "List")
		s = strings.TrimPrefix(s, "List<")
		s = strings.TrimSuffix(s, ">")
	}

	// Add the final inner type (e.g., "int", "string", "MyType")
	s = strings.TrimSpace(s)
	if s != "" {
		result = append(result, s)
	}

	return result
}

func ConvertToAliasRepr(types []string, pkgName []string) []AliasRepr {
	var result []AliasRepr
	for i, t := range types {
		result = append(result, AliasRepr{
			PackageName: pkgName[i], // or derive if format is pkg.Type
			TypeName:    t,
		})
	}
	return result
}

func (cse *CSharpEmitter) PostVisitTypeAliasType(node ast.Expr, indent int) {
	str := cse.emitAsString(";\n\n", 0)
	cse.gir.stack = append(cse.gir.stack, str)
	cse.aliases[cse.gir.stack[2]] = Alias{
		PackageName:    cse.pkg.Name + ".Api",
		representation: ConvertToAliasRepr(ParseNestedTypes(cse.gir.stack[4]), []string{"", cse.pkg.Name + ".Api"}),
		UnderlyingType: cse.pkg.TypesInfo.Types[node].Type.String(),
	}
	cse.gir.stack = mergeStackElements("@@PreVisitTypeAliasName", cse.gir.stack)
	if len(cse.gir.stack) == 1 {
		// TODO emit to aliases
		//cse.emitToFileBuffer(cse.stack[len(cse.stack)-1], "")
		cse.gir.stack = cse.gir.stack[:len(cse.gir.stack)-1]
	}
	cse.buffer = false
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	str := cse.emitAsString("return ", indent)
	cse.gir.emitToFileBuffer(str, "")

	if len(node.Results) == 1 {
		tv := cse.pkg.TypesInfo.Types[node.Results[0]]
		//pos := cse.pkg.Fset.Position(node.Pos())
		//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
		if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
			if !cse.isTuple && tv.Type.String() != "func()" {
				cse.gir.emitToFileBuffer("(", "")
				str := cse.emitAsString(typeVal, 0)
				cse.gir.emitToFileBuffer(str, "")
				cse.gir.emitToFileBuffer(")", "")
			}
		}
	}
	if len(node.Results) > 1 {
		str := cse.emitAsString("(", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	if len(node.Results) > 1 {
		str := cse.emitAsString(")", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
	str := cse.emitAsString(";", 0)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
}

func (cse *CSharpEmitter) PreVisitCallExpr(node *ast.CallExpr, indent int) {
	cse.gir.emitToFileBuffer("", "@PreVisitCallExpr")
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitCallExpr(node *ast.CallExpr, indent int) {
	pointerAndPosition := SearchPointerReverse("@PreVisitCallExpr", cse.gir.pointerAndPositionVec)
	if pointerAndPosition != nil {
		str, _ := ExtractSubstring(pointerAndPosition.Position, cse.gir.fileBuffer)
		for _, t := range destTypes {
			matchStr := t + "("
			if strings.Contains(str, matchStr) {
				cse.gir.fileBuffer, _ = RewriteFileBuffer(cse.gir.fileBuffer, pointerAndPosition.Position, matchStr, "("+t+")(")
			}
		}
	}
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitDeclStmt(node *ast.DeclStmt, indent int) {
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitDeclStmt(node *ast.DeclStmt, indent int) {
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := cse.emitAsString("", indent)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}
func (cse *CSharpEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	str := cse.emitAsString(";", 0)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	str := cse.emitAsString(cse.assignmentToken+" ", indent+1)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cse.isTuple = false
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int) {
	cse.gir.emitToFileBuffer("", "@PreVisitAssignStmtRhsExpr")
}

func (cse *CSharpEmitter) PostVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int) {
	pointerAndPosition := SearchPointerReverse("@PreVisitAssignStmtRhsExpr", cse.gir.pointerAndPositionVec)
	rewritten := false
	if pointerAndPosition != nil {
		str, _ := ExtractSubstring(pointerAndPosition.Position, cse.gir.fileBuffer)
		for _, t := range destTypes {
			matchStr := t + "("
			if strings.Contains(str, matchStr) {
				cse.gir.fileBuffer, _ = RewriteFileBuffer(cse.gir.fileBuffer, pointerAndPosition.Position, matchStr, "("+t+")(")
				rewritten = true
			}
		}
	}
	if !rewritten {
		tv := cse.pkg.TypesInfo.Types[node]
		//pos := cse.pkg.Fset.Position(node.Pos())
		//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
		if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
			if !cse.isTuple && tv.Type.String() != "func()" {
				cse.gir.fileBuffer, _ = RewriteFileBuffer(cse.gir.fileBuffer, pointerAndPosition.Position, "", "("+typeVal+")")
			}
		}
	}
}

func (cse *CSharpEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", indent)
		cse.gir.emitToFileBuffer(str, "")
	}
}

func (cse *CSharpEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	assignmentToken := node.Tok.String()
	if assignmentToken == ":=" && len(node.Lhs) == 1 {
		str := cse.emitAsString("var ", indent)
		cse.gir.emitToFileBuffer(str, "")
	} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
		str := cse.emitAsString("var (", indent)
		cse.gir.emitToFileBuffer(str, "")
	} else if assignmentToken == "=" && len(node.Lhs) > 1 {
		str := cse.emitAsString("(", indent)
		cse.gir.emitToFileBuffer(str, "")
		cse.isTuple = true
	}
	if assignmentToken != "+=" {
		assignmentToken = "="
	}
	cse.assignmentToken = assignmentToken
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
		str := cse.emitAsString(")", indent)
		cse.gir.emitToFileBuffer(str, "")
	} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
		str := cse.emitAsString(")", indent)
		cse.gir.emitToFileBuffer(str, "")
	}
	cse.shouldGenerate = false

}

func (cse *CSharpEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := cse.emitAsString("[", 0)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}
func (cse *CSharpEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	str := cse.emitAsString("]", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	str := cse.emitAsString("(", 1)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}
func (cse *CSharpEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	str := cse.emitAsString(")", 1)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	str := cse.emitAsString(op.String()+" ", 1)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
}
func (cse *CSharpEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	str := cse.emitAsString(";", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitIfStmt(node *ast.IfStmt, indent int) {
	cse.shouldGenerate = true
}
func (cse *CSharpEmitter) PostVisitIfStmt(node *ast.IfStmt, indent int) {
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cse.emitAsString("if (", 1)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	str := cse.emitAsString(")\n", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	str := cse.emitAsString("for (", indent)
	cse.gir.emitToFileBuffer(str, "")
	cse.insideForPostCond = true
	cse.insideForPostCond = true
}

func (cse *CSharpEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	if node == nil {
		str := cse.emitAsString(";", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
}

func (cse *CSharpEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	if node != nil {
		cse.insideForPostCond = false
	}
	str := cse.emitAsString(")\n", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	str := cse.emitAsString("else", 1)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	str := cse.emitAsString(";", 0)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PostVisitForStmt(node *ast.ForStmt, indent int) {
	cse.insideForPostCond = false
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	str := cse.emitAsString("foreach (var ", indent)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	str := cse.emitAsString(" in ", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	str := cse.emitAsString(")\n", 0)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	str := cse.emitAsString(node.Tok.String(), 0)
	if !cse.insideForPostCond {
		str += cse.emitAsString(";", 0)
	}
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PreVisitCompositeLitType(node ast.Expr, indent int) {
	str := cse.emitAsString("new ", 0)
	cse.gir.emitToFileBuffer(str, "")
	cse.gir.emitToFileBuffer("", "@PreVisitCompositeLitType")
}

func (cse *CSharpEmitter) PostVisitCompositeLitType(node ast.Expr, indent int) {
	pointerAndPosition := SearchPointerReverse("@PreVisitCompositeLitType", cse.gir.pointerAndPositionVec)
	if pointerAndPosition != nil {
		// TODO not very effective
		// go through all aliases and check if the underlying type matches
		for aliasName, alias := range cse.aliases {
			if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node].Type.Underlying().String() {
				cse.gir.fileBuffer, _ = RewriteFileBufferBetween(cse.gir.fileBuffer, pointerAndPosition.Position, len(cse.gir.fileBuffer), aliasName)
			}
		}
	}
}

func (cse *CSharpEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cse.emitAsString("{", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	str := cse.emitAsString("}", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	if index > 0 {
		str := cse.emitAsString(", ", 0)
		cse.gir.emitToFileBuffer(str, "")
	}
}

func (cse *CSharpEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	str := cse.emitAsString("[", 0)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = false
}

func (cse *CSharpEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	str := cse.emitAsString("]", 0)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitSliceExprLow(node ast.Expr, indent int) {
	cse.gir.emitToFileBuffer("..", "")
}

func (cse *CSharpEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cse.emitAsString("(", indent)
	cse.gir.emitToFileBuffer(str, "")
}
func (cse *CSharpEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	str := cse.emitAsString("}", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	str := cse.emitAsString(")", 0)
	str += cse.emitAsString("=>", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := ""
	if index > 0 {
		str += cse.emitAsString(", ", 0)
	}
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	str := cse.emitAsString(" ", 0)
	str += cse.emitAsString(node.Names[0].Name, indent)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	str := cse.emitAsString("{\n", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitFuncLitTypeResults(node *ast.FieldList, indent int) {
	cse.shouldGenerate = false
}
func (cse *CSharpEmitter) PostVisitFuncLitTypeResults(node *ast.FieldList, indent int) {
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	str := cse.emitAsString("object", indent)
	cse.gir.stack = append(cse.gir.stack, str)
}

func (cse *CSharpEmitter) PostVisitInterfaceType(node *ast.InterfaceType, indent int) {
	// emit only if it's not a complex type
	if len(cse.gir.stack) == 1 {
		cse.gir.emitToFileBuffer(cse.gir.stack[len(cse.gir.stack)-1], "")
		cse.gir.stack = cse.gir.stack[:len(cse.gir.stack)-1]
	}
}

func (cse *CSharpEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	str := cse.emitAsString("= ", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cse.emitAsString("(", 0)
	str += cse.emitAsString(node.Op.String(), 0)
	cse.gir.emitToFileBuffer(str, "")
}
func (cse *CSharpEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	str := cse.emitAsString(")", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func trimBeforeChar(s string, ch byte) string {
	pos := strings.IndexByte(s, ch)
	if pos == -1 {
		return s // character not found
	}
	return s[pos+1:]
}

func (cse *CSharpEmitter) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	// TODO dummy implementation
	// not very well performed
	for constIdent, obj := range cse.pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		if con, ok := obj.(*types.Const); ok {
			if constIdent.Name != node.Name {
				continue
			}
			constType := con.Type().String()
			constType = strings.TrimPrefix(constType, "untyped ")
			if constType == cse.pkg.TypesInfo.Defs[node].Type().String() {
				constType = trimBeforeChar(constType, '.')
			}
			str := cse.emitAsString(fmt.Sprintf("public const %s %s = ", constType, node.Name), 0)

			cse.gir.emitToFileBuffer(str, "")
		}
	}
}
func (cse *CSharpEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	str := cse.emitAsString(";\n", 0)
	cse.gir.emitToFileBuffer(str, "")
}
func (cse *CSharpEmitter) PostVisitGenDeclConst(node *ast.GenDecl, indent int) {
	str := cse.emitAsString("\n", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := cse.emitAsString("switch (", indent)
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}
func (cse *CSharpEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	str := cse.emitAsString("}", indent)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	str := cse.emitAsString(") {\n", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	cse.gir.emitToFileBuffer("\n", "")
	str := cse.emitAsString("break;\n", indent+4)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	if len(node) == 0 {
		str := cse.emitAsString("default:\n", indent+2)
		cse.gir.emitToFileBuffer(str, "")
	}
}

func (cse *CSharpEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cse.emitAsString("case ", indent+2)
	tv := cse.pkg.TypesInfo.Types[node]
	if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
		str += "(" + typeVal + ")"
	}
	cse.gir.emitToFileBuffer(str, "")
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	str := cse.emitAsString(":\n", 0)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := cse.emitAsString("(", indent)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PostVisitTypeAssertExprType(node ast.Expr, indent int) {
	str := cse.emitAsString(")", indent)
	cse.gir.emitToFileBuffer(str, "")
}

func (cse *CSharpEmitter) PreVisitKeyValueExpr(node *ast.KeyValueExpr, indent int) {
	cse.shouldGenerate = true
}

func (cse *CSharpEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	str := cse.emitAsString(node.Tok.String()+";", indent)
	cse.gir.emitToFileBuffer(str, "")
}
