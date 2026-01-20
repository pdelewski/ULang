package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
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
	Output      string
	OutputDir   string
	OutputName  string
	LinkRuntime string // Path to runtime directory (empty = disabled)
	file        *os.File
	BaseEmitter
	pkg               *packages.Package
	insideForPostCond bool
	assignmentToken   string
	forwardDecls      bool
	shouldGenerate    bool
	numFuncResults    int
	aliases           map[string]Alias
	currentPackage    string
	isArray           bool
	arrayType         string
	isTuple           bool
	isInfiniteLoop    bool // Track if current for loop is infinite (no init, cond, post)
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

// Helper function to determine token type for C# specific content
func (cse *CSharpEmitter) getTokenType(content string) TokenType {
	// Check for C# keywords
	switch content {
	case "using", "namespace", "class", "public", "private", "protected", "static", "override", "virtual", "sealed", "readonly", "var":
		return CSharpKeyword
	case "if", "else", "for", "while", "switch", "case", "default", "break", "continue", "return":
		return IfKeyword // Will be refined based on actual keyword
	case "(":
		return LeftParen
	case ")":
		return RightParen
	case "{":
		return LeftBrace
	case "}":
		return RightBrace
	case "[":
		return LeftBracket
	case "]":
		return RightBracket
	case ";":
		return Semicolon
	case ",":
		return Comma
	case ".":
		return Dot
	case "=", "+=", "-=", "*=", "/=":
		return Assignment
	case "+", "-", "*", "/", "%":
		return ArithmeticOperator
	case "==", "!=", "<", ">", "<=", ">=":
		return ComparisonOperator
	case "&&", "||", "!":
		return LogicalOperator
	case " ", "\t":
		return WhiteSpace
	case "\n":
		return NewLine
	}

	// Check if it's a number
	if len(content) > 0 && (content[0] >= '0' && content[0] <= '9') {
		return NumberLiteral
	}

	// Check if it's a string literal
	if len(content) >= 2 && content[0] == '"' && content[len(content)-1] == '"' {
		return StringLiteral
	}

	// Default to identifier
	return Identifier
}

// Helper function to emit token
func (cse *CSharpEmitter) emitToken(content string, tokenType TokenType, indent int) {
	token := CreateToken(tokenType, cse.emitAsString(content, indent))
	_ = cse.gir.emitTokenToFileBuffer(token, EmptyVisitMethod)
}
func (cse *CSharpEmitter) SetFile(file *os.File) {
	cse.file = file
}

func (cse *CSharpEmitter) GetFile() *os.File {
	return cse.file
}

func (cse *CSharpEmitter) executeIfNotForwardDecls(fn func()) {
	if cse.forwardDecls {
		return
	}
	fn()
}

func (cse *CSharpEmitter) PreVisitProgram(indent int) {
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
	_ = cse.gir.emitToFileBuffer(str, EmptyVisitMethod)

	cse.insideForPostCond = false
}

func (cse *CSharpEmitter) PostVisitProgram(indent int) {
	emitTokensToFile(cse.file, cse.gir.tokenSlice)
	cse.file.Close()

	// Generate .NET project files if link-runtime is enabled
	if cse.LinkRuntime != "" {
		if err := cse.GenerateCsproj(); err != nil {
			log.Printf("Warning: %v", err)
		}
		if err := cse.CopyGraphicsRuntime(); err != nil {
			log.Printf("Warning: %v", err)
		}
	}
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatures(indent int) {
	cse.forwardDecls = true
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignatures(indent int) {
	cse.forwardDecls = false
}

func (cse *CSharpEmitter) PreVisitFuncDeclName(node *ast.Ident, indent int) {
	cse.executeIfNotForwardDecls(func() {
		var str string
		if node.Name == "main" {
			str = cse.emitAsString(fmt.Sprintf("Main"), 0)
		} else {
			str = cse.emitAsString(fmt.Sprintf("%s", node.Name), 0)
		}
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitBlockStmt(node *ast.BlockStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("{", LeftBrace, 1)
		str := cse.emitAsString("\n", 1)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitBlockStmt(node *ast.BlockStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("}", RightBrace, 1)
		cse.isArray = false
	})
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("(", LeftParen, 0)
	})
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignatureTypeParams(node *ast.FuncDecl, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, 0)
	})
}

func (cse *CSharpEmitter) PreVisitIdent(e *ast.Ident, indent int) {
	cse.executeIfNotForwardDecls(func() {
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

		cse.emitToken(str, Identifier, 0)
	})
}

func (cse *CSharpEmitter) PreVisitCallExprArgs(node []ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("(", LeftParen, 0)
	})
}

func (cse *CSharpEmitter) PostVisitCallExprArgs(node []ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, 0)
	})
}

func (cse *CSharpEmitter) PreVisitBasicLit(e *ast.BasicLit, indent int) {
	cse.executeIfNotForwardDecls(func() {
		var str string
		if e.Kind == token.STRING {
			e.Value = strings.Replace(e.Value, "\"", "", -1)
			if len(e.Value) > 0 && e.Value[0] == '`' {
				e.Value = strings.Replace(e.Value, "`", "", -1)
				str = (cse.emitAsString(fmt.Sprintf("@\"(%s)\"", e.Value), 0))
			} else {
				str = (cse.emitAsString(fmt.Sprintf("@\"%s\"", e.Value), 0))
			}
			cse.emitToken(str, StringLiteral, 0)
		} else {
			str = (cse.emitAsString(e.Value, 0))
			cse.emitToken(str, NumberLiteral, 0)
		}
	})
}

func (cse *CSharpEmitter) PreVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		// Reset isArray flag at the start of each variable declaration
		// This prevents stale state from previous declarations affecting this one
		cse.isArray = false
	})
}

func (cse *CSharpEmitter) PostVisitDeclStmtValueSpecType(node *ast.ValueSpec, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		pointerAndPosition := SearchPointerIndexReverse(PreVisitDeclStmtValueSpecType, cse.gir.pointerAndIndexVec)
		if pointerAndPosition != nil {
			for aliasName, alias := range cse.aliases {
				if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
					cse.gir.tokenSlice, _ = RewriteTokensBetween(cse.gir.tokenSlice, pointerAndPosition.Index, len(cse.gir.tokenSlice), []string{aliasName})
				}
			}
		}
	})
}

func (cse *CSharpEmitter) PreVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(" ", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitDeclStmtValueSpecNames(node *ast.Ident, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		var str string
		if cse.isArray {
			str += " = new "
			str += strings.TrimSpace(cse.arrayType)
			str += "();"
			cse.isArray = false
		} else {
			str += " = default;"
		}
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitGenStructFieldType(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("public", indent+2)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitGenStructFieldType(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.gir.emitToFileBuffer(" ", EmptyVisitMethod)
		// clean array marker as we should generate
		// initializer only for expression statements
		// not for struct fields
		cse.isArray = false
	})
}

func (cse *CSharpEmitter) PostVisitGenStructFieldName(node *ast.Ident, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.gir.emitToFileBuffer(";\n", EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitPackage(pkg *packages.Package, indent int) {
	cse.executeIfNotForwardDecls(func() {
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
		err := cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		err = cse.gir.emitToFileBufferString("", pkg.Name)
		cse.currentPackage = packageName
		str = cse.emitAsString(fmt.Sprintf("public struct %s {\n\n", "Api"), indent+2)
		err = cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	})
}

func (cse *CSharpEmitter) PostVisitPackage(pkg *packages.Package, indent int) {
	cse.executeIfNotForwardDecls(func() {
		pointerAndPosition := SearchPointerIndexReverseString(pkg.Name, cse.gir.pointerAndIndexVec)
		if pointerAndPosition != nil {
			var newStr string
			for aliasKey, aliasVal := range cse.aliases {
				aliasRepr := RebuildNestedType(aliasVal.representation)
				newStr += "using " + aliasKey + " = " + aliasRepr + ";\n"
			}
			newStr += "\n"
			cse.gir.tokenSlice, _ = RewriteTokens(cse.gir.tokenSlice, pointerAndPosition.Index, []string{}, []string{newStr})
		}

		str := cse.emitAsString("}\n", indent+2)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		err := cse.gir.emitToFileBuffer("}\n", EmptyVisitMethod)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	})
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignature(node *ast.FuncDecl, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.isArray = false
	})
}

func (cse *CSharpEmitter) PostVisitBlockStmtList(node ast.Stmt, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("\n", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitFuncDecl(node *ast.FuncDecl, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("\n\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitGenStructInfo(node GenTypeInfo, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(fmt.Sprintf("public struct %s\n", node.Name), indent+2)
		str += cse.emitAsString("{\n", indent+2)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitGenStructInfo(node GenTypeInfo, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("};\n\n", indent+2)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitArrayType(node ast.ArrayType, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("List", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		str = cse.emitAsString("<", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}
func (cse *CSharpEmitter) PostVisitArrayType(node ast.ArrayType, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(">", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)

		pointerAndPosition := SearchPointerIndexReverse(PreVisitArrayType, cse.gir.pointerAndIndexVec)
		if pointerAndPosition != nil {
			tokens, _ := ExtractTokens(pointerAndPosition.Index, cse.gir.tokenSlice)
			cse.isArray = true
			cse.arrayType = strings.Join(tokens, "")
		}
	})
}

func (cse *CSharpEmitter) PreVisitFuncType(node *ast.FuncType, indent int) {
	cse.executeIfNotForwardDecls(func() {
		var str string
		if node.Results != nil {
			str = cse.emitAsString("Func", indent)
		} else {
			str = cse.emitAsString("Action", indent)
		}
		str += cse.emitAsString("<", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}
func (cse *CSharpEmitter) PostVisitFuncType(node *ast.FuncType, indent int) {
	cse.executeIfNotForwardDecls(func() {
		pointerAndPosition := SearchPointerIndexReverse(PreVisitFuncType, cse.gir.pointerAndIndexVec)
		if pointerAndPosition != nil && cse.numFuncResults > 0 {
			// For function types with return values, we need to reorder tokens
			// to move return type to the end (C# syntax requirement)
			tokens, _ := ExtractTokens(pointerAndPosition.Index, cse.gir.tokenSlice)
			if len(tokens) > 2 {
				// Find and move return type to end with comma separator
				var reorderedTokens []string
				reorderedTokens = append(reorderedTokens, tokens[0]) // "Func<" or "Action<"
				if len(tokens) > 3 {
					// Skip return type (index 1) and add parameters first
					reorderedTokens = append(reorderedTokens, tokens[2:]...)
					reorderedTokens = append(reorderedTokens, ",")
					reorderedTokens = append(reorderedTokens, tokens[1]) // Add return type at end
				}
				cse.gir.tokenSlice, _ = RewriteTokensBetween(cse.gir.tokenSlice, pointerAndPosition.Index, len(cse.gir.tokenSlice), reorderedTokens)
			}
		}

		str := cse.emitAsString(">", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitFuncTypeParam(node *ast.Field, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if index > 0 {
			str := cse.emitAsString(", ", 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PostVisitSelectorExprX(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
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
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitFuncTypeResults(node *ast.FieldList, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if node != nil {
			cse.numFuncResults = len(node.List)
		}
	})
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsList(node *ast.Field, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if index > 0 {
			str := cse.emitAsString(", ", 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeParamsArgName(node *ast.Ident, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.gir.emitToFileBuffer(" ", EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if index > 0 {
			cse.emitToken(",", Comma, 0)
		}
	})
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignatureTypeResultsList(node *ast.Field, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		pointerAndPosition := SearchPointerIndexReverse(PreVisitFuncDeclSignatureTypeResultsList, cse.gir.pointerAndIndexVec)
		if pointerAndPosition != nil {
			adjustment := 0
			// Check for comma after the type to adjust index
			if cse.gir.tokenSlice[pointerAndPosition.Index].Content == "," {
				adjustment = 1
			}
			for aliasName, alias := range cse.aliases {
				if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node.Type].Type.Underlying().String() {
					cse.gir.tokenSlice, _ = RewriteTokensBetween(cse.gir.tokenSlice, pointerAndPosition.Index+adjustment, len(cse.gir.tokenSlice), []string{aliasName})
				}
			}
		}
	})
}

func (cse *CSharpEmitter) PreVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("public static ", indent+2)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		if node.Type.Results != nil {
			if len(node.Type.Results.List) > 1 {
				cse.emitToken("(", LeftParen, 0)
			}
		} else {
			str := cse.emitAsString("void", 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PostVisitFuncDeclSignatureTypeResults(node *ast.FuncDecl, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if node.Type.Results != nil {
			if len(node.Type.Results.List) > 1 {
				cse.emitToken(")", RightParen, 0)
			}
		}

		str := cse.emitAsString("", 1)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitTypeAliasName(node *ast.Ident, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("using ", indent+2)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitTypeAliasName(node *ast.Ident, indent int) {
	cse.executeIfNotForwardDecls(func() {
	})
}

func (cse *CSharpEmitter) PreVisitTypeAliasType(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.gir.emitToFileBuffer(" = ", EmptyVisitMethod)
	})
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

func (cse *CSharpEmitter) PostVisitTypeAliasType(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		// Extract tokens for alias processing
		pointerAndPosition := SearchPointerIndexReverse(PreVisitTypeAliasName, cse.gir.pointerAndIndexVec)
		if pointerAndPosition != nil {
			tokens, _ := ExtractTokens(pointerAndPosition.Index, cse.gir.tokenSlice)
			if len(tokens) >= 3 {
				// tokens[0] = "using ", tokens[1] = alias name, tokens[2] = " = ", tokens[3+] = type
				aliasName := tokens[1]
				typeTokens := tokens[3:]
				typeStr := strings.Join(typeTokens, "")
				cse.aliases[aliasName] = Alias{
					PackageName:    cse.pkg.Name + ".Api",
					representation: ConvertToAliasRepr(ParseNestedTypes(typeStr), []string{"", cse.pkg.Name + ".Api"}),
					UnderlyingType: cse.pkg.TypesInfo.Types[node].Type.String(),
				}
			}
			// Remove the alias declaration from the current position - it will be added at the top later
			cse.gir.tokenSlice, _ = RewriteTokensBetween(cse.gir.tokenSlice, pointerAndPosition.Index, len(cse.gir.tokenSlice), []string{})
		}
	})
}

func (cse *CSharpEmitter) PreVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("return ", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)

		if len(node.Results) == 1 {
			tv := cse.pkg.TypesInfo.Types[node.Results[0]]
			//pos := cse.pkg.Fset.Position(node.Pos())
			//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
			if tv.Type != nil {
				if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
					if !cse.isTuple && tv.Type.String() != "func()" {
						cse.emitToken("(", LeftParen, 0)
						str := cse.emitAsString(typeVal, 0)
						cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
						cse.emitToken(")", RightParen, 0)
					}
				}
			}
		}
		if len(node.Results) > 1 {
			cse.emitToken("(", LeftParen, 0)
		}
	})
}

func (cse *CSharpEmitter) PostVisitReturnStmt(node *ast.ReturnStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if len(node.Results) > 1 {
			cse.emitToken(")", RightParen, 0)
		}
		str := cse.emitAsString(";", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitReturnStmtResult(node ast.Expr, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if index > 0 {
			str := cse.emitAsString(", ", 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PostVisitCallExpr(node *ast.CallExpr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		pointerAndPosition := SearchPointerIndexReverse(PreVisitCallExpr, cse.gir.pointerAndIndexVec)
		if pointerAndPosition != nil {
			tokens, _ := ExtractTokens(pointerAndPosition.Index, cse.gir.tokenSlice)
			for _, t := range destTypes {
				if len(tokens) >= 2 && tokens[0] == t && tokens[1] == "(" {
					cse.gir.tokenSlice, _ = RewriteTokens(cse.gir.tokenSlice, pointerAndPosition.Index, []string{tokens[0], tokens[1]}, []string{"(", t, ")", "("})
				}
			}
		}
	})
}

func (cse *CSharpEmitter) PreVisitAssignStmt(node *ast.AssignStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitAssignStmt(node *ast.AssignStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		// Don't emit semicolon inside for loop post statement
		if !cse.insideForPostCond {
			str := cse.emitAsString(";", 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PreVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		opTokenType := cse.getTokenType(cse.assignmentToken)
		cse.emitToken(cse.assignmentToken, opTokenType, indent+1)
		cse.emitToken(" ", WhiteSpace, 0)
	})
}

func (cse *CSharpEmitter) PostVisitAssignStmtRhs(node *ast.AssignStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.isTuple = false
	})
}

func (cse *CSharpEmitter) PostVisitAssignStmtRhsExpr(node ast.Expr, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		pointerAndPosition := SearchPointerIndexReverse(PreVisitAssignStmtRhsExpr, cse.gir.pointerAndIndexVec)
		rewritten := false
		if pointerAndPosition != nil {
			tokens, _ := ExtractTokens(pointerAndPosition.Index, cse.gir.tokenSlice)
			for _, t := range destTypes {
				if len(tokens) >= 2 && tokens[0] == t && tokens[1] == "(" {
					cse.gir.tokenSlice, _ = RewriteTokens(cse.gir.tokenSlice, pointerAndPosition.Index, []string{tokens[0], tokens[1]}, []string{"(", t, ")", "("})
				}
			}
		}

		if !rewritten {
			tv := cse.pkg.TypesInfo.Types[node]
			//pos := cse.pkg.Fset.Position(node.Pos())
			//fmt.Printf("@@Type: %s %s:%d:%d\n", tv.Type, pos.Filename, pos.Line, pos.Column)
			if tv.Type != nil {
				if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
					if !cse.isTuple && tv.Type.String() != "func()" {
						cse.gir.tokenSlice, _ = RewriteTokens(cse.gir.tokenSlice, pointerAndPosition.Index, []string{}, []string{"(", typeVal, ")"})
					}
				}
			}
		}
	})
}

func (cse *CSharpEmitter) PreVisitAssignStmtLhsExpr(node ast.Expr, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if index > 0 {
			str := cse.emitAsString(", ", indent)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PreVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		assignmentToken := node.Tok.String()
		if assignmentToken == ":=" && len(node.Lhs) == 1 {
			str := cse.emitAsString("var ", indent)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		} else if assignmentToken == ":=" && len(node.Lhs) > 1 {
			str := cse.emitAsString("var ", indent)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
			cse.emitToken("(", LeftParen, 0)
		} else if assignmentToken == "=" && len(node.Lhs) > 1 {
			cse.emitToken("(", LeftParen, indent)
			cse.isTuple = true
		}
		// Preserve compound assignment operators, convert := to =
		if assignmentToken != "+=" && assignmentToken != "-=" && assignmentToken != "*=" && assignmentToken != "/=" {
			assignmentToken = "="
		}
		cse.assignmentToken = assignmentToken
	})
}

func (cse *CSharpEmitter) PostVisitAssignStmtLhs(node *ast.AssignStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if node.Tok.String() == ":=" && len(node.Lhs) > 1 {
			cse.emitToken(")", RightParen, indent)
		} else if node.Tok.String() == "=" && len(node.Lhs) > 1 {
			cse.emitToken(")", RightParen, indent)
		}
	})
}

func (cse *CSharpEmitter) PreVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("[", LeftBracket, 0)
	})
}
func (cse *CSharpEmitter) PostVisitIndexExprIndex(node *ast.IndexExpr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("]", RightBracket, 0)
	})
}

func (cse *CSharpEmitter) PreVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("(", LeftParen, 1)
	})
}
func (cse *CSharpEmitter) PostVisitBinaryExpr(node *ast.BinaryExpr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, 1)
	})
}

func (cse *CSharpEmitter) PreVisitBinaryExprOperator(op token.Token, indent int) {
	cse.executeIfNotForwardDecls(func() {
		opTokenType := cse.getTokenType(op.String())
		cse.emitToken(op.String(), opTokenType, 1)
		cse.emitToken(" ", WhiteSpace, 0)
	})
}

func (cse *CSharpEmitter) PreVisitCallExprArg(node ast.Expr, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if index > 0 {
			str := cse.emitAsString(", ", 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}
func (cse *CSharpEmitter) PostVisitExprStmtX(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(";", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitIfStmtCond(node *ast.IfStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("if ", 1)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		cse.emitToken("(", LeftParen, 0)
	})
}

func (cse *CSharpEmitter) PostVisitIfStmtCond(node *ast.IfStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, 0)
		str := cse.emitAsString("\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitForStmt(node *ast.ForStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("for ", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		cse.emitToken("(", LeftParen, 0)
	})
}

func (cse *CSharpEmitter) PostVisitForStmtInit(node ast.Stmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if node == nil {
			str := cse.emitAsString(";", 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PreVisitForStmtPost(node ast.Stmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if node != nil {
			cse.insideForPostCond = true
		}
	})
}

func (cse *CSharpEmitter) PostVisitForStmtPost(node ast.Stmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.insideForPostCond = false
		cse.emitToken(")", RightParen, 0)
		str := cse.emitAsString("\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitIfStmtElse(node *ast.IfStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("else", 1)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitForStmtCond(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(";", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitForStmt(node *ast.ForStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.insideForPostCond = false
	})
}

func (cse *CSharpEmitter) PreVisitRangeStmt(node *ast.RangeStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("foreach ", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		cse.emitToken("(", LeftParen, 0)
		str = cse.emitAsString("var ", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitRangeStmtValue(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(" in ", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitRangeStmtX(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, 0)
		str := cse.emitAsString("\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitIncDecStmt(node *ast.IncDecStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(node.Tok.String(), 0)
		if !cse.insideForPostCond {
			str += cse.emitAsString(";", 0)
		}
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitCompositeLitType(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("new ", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitCompositeLitType(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		pointerAndPosition := SearchPointerIndexReverse(PreVisitCompositeLitType, cse.gir.pointerAndIndexVec)
		if pointerAndPosition != nil {
			// TODO not very effective
			// go through all aliases and check if the underlying type matches
			for aliasName, alias := range cse.aliases {
				if alias.UnderlyingType == cse.pkg.TypesInfo.Types[node].Type.Underlying().String() {
					const newKeywordIndex = 1
					cse.gir.tokenSlice, _ = RewriteTokensBetween(cse.gir.tokenSlice, pointerAndPosition.Index+newKeywordIndex, len(cse.gir.tokenSlice), []string{aliasName})
				}
			}
		}
	})
}

func (cse *CSharpEmitter) PreVisitCompositeLitElts(node []ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("{", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitCompositeLitElts(node []ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("}", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitCompositeLitElt(node ast.Expr, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if index > 0 {
			str := cse.emitAsString(", ", 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PostVisitSliceExprX(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("[", LeftBracket, 0)
	})
}

func (cse *CSharpEmitter) PostVisitSliceExpr(node *ast.SliceExpr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("]", RightBracket, 0)
	})
}

func (cse *CSharpEmitter) PostVisitSliceExprLow(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.gir.emitToFileBuffer("..", EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitFuncLit(node *ast.FuncLit, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("(", LeftParen, indent)
	})
}
func (cse *CSharpEmitter) PostVisitFuncLit(node *ast.FuncLit, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("}", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitFuncLitTypeParams(node *ast.FieldList, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, 0)
		str := cse.emitAsString("=>", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := ""
		if index > 0 {
			str += cse.emitAsString(", ", 0)
		}
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitFuncLitTypeParam(node *ast.Field, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(" ", 0)
		str += cse.emitAsString(node.Names[0].Name, indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitFuncLitBody(node *ast.BlockStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("{\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitFuncLitTypeResult(node *ast.Field, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.shouldGenerate = false
	})
}

func (cse *CSharpEmitter) PostVisitFuncLitTypeResult(node *ast.Field, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.shouldGenerate = true
	})
}

func (cse *CSharpEmitter) PreVisitInterfaceType(node *ast.InterfaceType, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("object", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitKeyValueExprValue(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("= ", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("(", LeftParen, 0)
		str := cse.emitAsString(node.Op.String(), 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}
func (cse *CSharpEmitter) PostVisitUnaryExpr(node *ast.UnaryExpr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, 0)
	})
}

func trimBeforeChar(s string, ch byte) string {
	pos := strings.IndexByte(s, ch)
	if pos == -1 {
		return s // character not found
	}
	return s[pos+1:]
}

func (cse *CSharpEmitter) PreVisitGenDeclConstName(node *ast.Ident, indent int) {
	cse.executeIfNotForwardDecls(func() {
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

				cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
			}
		}
	})
}
func (cse *CSharpEmitter) PostVisitGenDeclConstName(node *ast.Ident, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(";\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}
func (cse *CSharpEmitter) PostVisitGenDeclConst(node *ast.GenDecl, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitSliceExprXBegin(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.shouldGenerate = false
	})
}

func (cse *CSharpEmitter) PostVisitSliceExprXBegin(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.shouldGenerate = true
	})
}

func (cse *CSharpEmitter) PreVisitSliceExprXEnd(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.shouldGenerate = false
	})
}

func (cse *CSharpEmitter) PostVisitSliceExprXEnd(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.shouldGenerate = true
	})
}

func (cse *CSharpEmitter) PreVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("switch ", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		cse.emitToken("(", LeftParen, 0)
	})
}
func (cse *CSharpEmitter) PostVisitSwitchStmt(node *ast.SwitchStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("}", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitSwitchStmtTag(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, 0)
		str := cse.emitAsString(" ", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		str = cse.emitAsString("{\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitCaseClause(node *ast.CaseClause, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.gir.emitToFileBuffer("\n", EmptyVisitMethod)
		str := cse.emitAsString("break;\n", indent+4)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PostVisitCaseClauseList(node []ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		if len(node) == 0 {
			str := cse.emitAsString("default:\n", indent+2)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		}
	})
}

func (cse *CSharpEmitter) PreVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString("case ", indent+2)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
		tv := cse.pkg.TypesInfo.Types[node]
		if typeVal, ok := csTypesMap[tv.Type.String()]; ok {
			cse.emitToken("(", LeftParen, 0)
			str = cse.emitAsString(typeVal, 0)
			cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
			cse.emitToken(")", RightParen, 0)
		}
	})
}

func (cse *CSharpEmitter) PostVisitCaseClauseListExpr(node ast.Expr, index int, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(":\n", 0)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

func (cse *CSharpEmitter) PreVisitTypeAssertExprType(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken("(", LeftParen, indent)
	})
}

func (cse *CSharpEmitter) PostVisitTypeAssertExprType(node ast.Expr, indent int) {
	cse.executeIfNotForwardDecls(func() {
		cse.emitToken(")", RightParen, indent)
	})
}

func (cse *CSharpEmitter) PreVisitBranchStmt(node *ast.BranchStmt, indent int) {
	cse.executeIfNotForwardDecls(func() {
		str := cse.emitAsString(node.Tok.String()+";", indent)
		cse.gir.emitToFileBuffer(str, EmptyVisitMethod)
	})
}

// GenerateCsproj creates a .csproj file for building the C# project with SDL2-CS
func (cse *CSharpEmitter) GenerateCsproj() error {
	if cse.LinkRuntime == "" {
		return nil
	}

	csprojPath := filepath.Join(cse.OutputDir, cse.OutputName+".csproj")
	file, err := os.Create(csprojPath)
	if err != nil {
		return fmt.Errorf("failed to create .csproj: %w", err)
	}
	defer file.Close()

	csproj := `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net9.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
    <AllowUnsafeBlocks>true</AllowUnsafeBlocks>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Sayers.SDL2.Core" Version="1.0.11" />
  </ItemGroup>

  <!-- Copy native SDL2 library on macOS (Homebrew installation) -->
  <Target Name="CopyNativeSDL2" AfterTargets="Build" Condition="$([MSBuild]::IsOSPlatform('OSX'))">
    <PropertyGroup>
      <SDL2HomebrewPath Condition="Exists('/opt/homebrew/lib/libSDL2.dylib')">/opt/homebrew/lib/libSDL2.dylib</SDL2HomebrewPath>
      <SDL2HomebrewPath Condition="$(SDL2HomebrewPath) == '' And Exists('/usr/local/lib/libSDL2.dylib')">/usr/local/lib/libSDL2.dylib</SDL2HomebrewPath>
    </PropertyGroup>
    <Copy SourceFiles="$(SDL2HomebrewPath)" DestinationFolder="$(OutputPath)" Condition="$(SDL2HomebrewPath) != ''" />
    <Warning Text="SDL2 native library not found. Install with: brew install sdl2" Condition="$(SDL2HomebrewPath) == ''" />
  </Target>

</Project>
`

	_, err = file.WriteString(csproj)
	if err != nil {
		return fmt.Errorf("failed to write .csproj: %w", err)
	}

	DebugLogPrintf("Generated .csproj at %s", csprojPath)
	return nil
}

// CopyGraphicsRuntime copies the graphics runtime file from the runtime directory
func (cse *CSharpEmitter) CopyGraphicsRuntime() error {
	if cse.LinkRuntime == "" {
		return nil
	}

	// Source path: LinkRuntime points to runtime directory, graphics runtime is in graphics/csharp/
	runtimeSrcPath := filepath.Join(cse.LinkRuntime, "graphics", "csharp", "GraphicsRuntime.cs")
	graphicsCs, err := os.ReadFile(runtimeSrcPath)
	if err != nil {
		return fmt.Errorf("failed to read graphics runtime from %s: %w", runtimeSrcPath, err)
	}

	// Destination path
	graphicsPath := filepath.Join(cse.OutputDir, "GraphicsRuntime.cs")
	if err := os.WriteFile(graphicsPath, graphicsCs, 0644); err != nil {
		return fmt.Errorf("failed to write GraphicsRuntime.cs: %w", err)
	}

	DebugLogPrintf("Copied GraphicsRuntime.cs from %s to %s", runtimeSrcPath, graphicsPath)
	return nil
}
