using System;
using System.Collections;
using System.Collections.Generic;

public static class SliceBuiltins {
    public static List<T> Append<T>(this List<T> list, T element)
    {
        var result = new List<T>(list);
        result.Add(element);
        return result;
    }

    public static List<T> Append<T>(this List<T> list, params T[] elements)
    {
        var result = new List<T>(list);
        result.AddRange(elements);
        return result;
    }

    public static List<T> Append<T>(this List<T> list, List<T> elements)
    {
        var result = new List<T>(list);
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
namespace lexer {

public class Api {

    public class Token {
        public  sbyte Type;
        public  List<sbyte> Representation;
    };

    public const int TokenTypeIdentifier = 1;
    public const int TokenTypeOperator = 2;
    public const int TokenTypeNumber = 3;
    public const int TokenTypeWhitespace = 4;
    public const int TokenTypeDot = 5;
    public const int TokenTypeSemicolon = 6;

    public const int TokenLetter = 0;
    public const int TokenDigit = 1;
    public const int TokenSpace = 2;
    public const int TokenSymbol = 3;
    public const int TokenLeftParenthesis = 4;
    public const int TokenRightParenthesis = 5;
    public const int TokenPipe = 6;
    public const int TokenGreater = 7;
    public const int TokenLess = 8;

    public static bool IsDigit(sbyte b)
    {
        return  ( (b >= '0' ) &&  (b <= '9' ) );
    }

    public static bool IsAlpha(sbyte b)
    {
        return  ( ( (b >= 'a' ) &&  (b <= 'z' ) ) ||  ( ( (b >= 'A' ) &&  (b <= 'Z' ) ) ||  (b == '_' ) ) );
    }

    public static bool IsEqual(sbyte b)
    {
        return  (b == '=' );
    }

    public static bool IsSemicolon(sbyte b)
    {
        return  (b == ';' );
    }

    public static bool IsFrom(Token token)
    {
        return  ( ( ( ( (SliceBuiltins.Length(token.Representation) == 4 ) &&  (token.Representation[0] == 'f' ) ) &&  (token.Representation[1] == 'r' ) ) &&  (token.Representation[2] == 'o' ) ) &&  (token.Representation[3] == 'm' ) );
    }

    public static bool IsSelect(Token token)
    {
        return  ( ( ( ( ( ( (SliceBuiltins.Length(token.Representation) == 6 ) &&  (token.Representation[0] == 's' ) ) &&  (token.Representation[1] == 'e' ) ) &&  (token.Representation[2] == 'l' ) ) &&  (token.Representation[3] == 'e' ) ) &&  (token.Representation[4] == 'c' ) ) &&  (token.Representation[5] == 't' ) );
    }

    public static bool IsWhere(Token token)
    {
        return  ( ( ( ( ( (SliceBuiltins.Length(token.Representation) == 5 ) &&  (token.Representation[0] == 'w' ) ) &&  (token.Representation[1] == 'h' ) ) &&  (token.Representation[2] == 'e' ) ) &&  (token.Representation[3] == 'r' ) ) &&  (token.Representation[4] == 'e' ) );
    }

    public static void DumpToken(Token token)
    {
        Formatter.Printf(@"Token type: %d ", token.Type);
        foreach (var b in token.Representation) {
            if ( (b == ' ' )) {
                Formatter.Printf(@" ");
            } else if ( (b == '\t' )) {
                Formatter.Printf(@"\t");
            } else if ( (b == '\n' )) {
                Formatter.Printf(@"\n");
            } else if ( (b == '.' )) {
                Formatter.Printf(@".");
            } else {
                Formatter.Printf(@"%c", b);
            }
        }
        Console.WriteLine();
    }

    public static void DumpTokens(List<Token> tokens)
    {
        foreach (var token in tokens) {
            DumpToken(token);
        }
    }

    public static string DumpTokenString(Token token)
    {
        string result = default;
        result += (string)Formatter.Sprintf(@"Token type: %d ", token.Type);
        foreach (var b in token.Representation) {
            switch (b) {
            case (sbyte)' ':
                result += (string)@" ";
                break;
            case (sbyte)'\t':
                result += (string)@"\t";
                break;
            case (sbyte)'\n':
                result += (string)@"\n";
                break;
            case (sbyte)'.':
                result += (string)@".";
                break;
            default:
                result += (string)Formatter.Sprintf(@"%c", b);
                break;
            }
        }
        result += (string)@"\n";
        return (string)result;
    }

    public static string DumpTokensString(List<Token> tokens)
    {
        string result = default;
        foreach (var token in tokens) {
            result += (string)Formatter.Sprintf(@"Token type: %d ", token.Type);
            foreach (var b in token.Representation) {
                switch (b) {
                case (sbyte)' ':
                    result += (string)@" ";
                    break;
                case (sbyte)'\t':
                    result += (string)@"\t";
                    break;
                case (sbyte)'\n':
                    result += (string)@"\n";
                    break;
                case (sbyte)'.':
                    result += (string)@".";
                    break;
                default:
                    result += (string)Formatter.Sprintf(@"%c", b);
                    break;
                }
            }
            result += (string)@"\n";
        }
        return (string)result;
    }

    public static List<Token> GetTokens(Token token)
    {
        List<Token> tokens = new List<Token>();
        Token currentToken = new Token {
            Representation = new List<sbyte>()
        };
        foreach (var b in token.Representation) {
            if ( (b == ';' )) {
                if ( (SliceBuiltins.Length(currentToken.Representation) > 0 )) {
                    tokens = SliceBuiltins.Append(tokens, currentToken);
                    currentToken = new Token {
                        Representation = new List<sbyte>()
                    };
                }
                tokens = SliceBuiltins.Append(tokens, new Token {Type= TokenTypeSemicolon, Representation= new List<sbyte>{b}});
                continue;
            }
            if ( ( ( (b == ' ' ) ||  (b == '\t' ) ) ||  (b == '\n' ) )) {
                if ( (SliceBuiltins.Length(currentToken.Representation) > 0 )) {
                    tokens = SliceBuiltins.Append(tokens, currentToken);
                    currentToken = new Token {
                        Representation = new List<sbyte>()
                    };
                }
            } else {
                currentToken.Type = (sbyte)TokenTypeIdentifier;
                currentToken.Representation = SliceBuiltins.Append(currentToken.Representation, b);
            }
        }
        if ( (SliceBuiltins.Length(currentToken.Representation) > 0 )) {
            tokens = SliceBuiltins.Append(tokens, currentToken);
        }
        return tokens;
    }

    public static (Token,List<Token>) GetNextToken(List<Token> tokens)
    {
        if ( (SliceBuiltins.Length(tokens) == 0 )) {
            return (new Token {
                Representation = new List<sbyte>()
            }, new List<Token> {});
        }
        return (tokens[0], tokens[1..]);
    }

    public static Token StringToToken(string s)
    {
        Token token = new Token();
        token.Representation = new List<sbyte>();
        foreach (var r in s) {
            token.Representation = SliceBuiltins.Append(token.Representation, (sbyte)(r));
        }
        return token;
    }

    public static bool IsLetter(sbyte b)
    {
        return  ( ( (b >= 'a' ) &&  (b <= 'z' ) ) ||  ( (b >= 'A' ) &&  (b <= 'Z' ) ) );
    }

    public static bool IsAlphaNumeric(sbyte b)
    {
        return  ( (IsLetter(b) || IsDigit(b) ) ||  (b == '_' ) );
    }

    public static bool IsSpace(sbyte b)
    {
        return  (b == ' ' );
    }

    public static bool IsLeftParenthesis(sbyte b)
    {
        return  (b == '(' );
    }

    public static bool IsRightParenthesis(sbyte b)
    {
        return  (b == ')' );
    }

    public static bool IsPipe(sbyte b)
    {
        return  (b == '|' );
    }

    public static bool IsGreater(sbyte b)
    {
        return  (b == '>' );
    }

    public static bool IsLess(sbyte b)
    {
        return  (b == '<' );
    }

    public static List<Token> Tokenize(string text)
    {
        List<Token> tokens = new List<Token>();
        List<sbyte> currentToken = new List<sbyte>();
        sbyte currentType = default;
        var addToken = ()=> {
            {
                if ( (SliceBuiltins.Length(currentToken) > 0 )) {
                    tokens = SliceBuiltins.Append(tokens, new Token {Type= currentType, Representation= currentToken});
                    currentToken = new List<sbyte>();
                }
            }
        };
        for (var i = 0; (i < SliceBuiltins.Length(text) ); i++) {
            var c = (sbyte)(sbyte)(text[i]);
            sbyte tokenType = default;
            if (IsAlphaNumeric(c)) {
                tokenType = (sbyte)TokenLetter;
            } else if (IsDigit(c)) {
                tokenType = (sbyte)TokenDigit;
            } else if (IsSpace(c)) {
                tokenType = (sbyte)TokenSpace;
            } else if (IsLeftParenthesis(c)) {
                tokenType = (sbyte)TokenLeftParenthesis;
            } else if (IsRightParenthesis(c)) {
                tokenType = (sbyte)TokenRightParenthesis;
            } else if (IsPipe(c)) {
                tokenType = (sbyte)TokenPipe;
            } else if (IsGreater(c)) {
                tokenType = (sbyte)TokenGreater;
            } else if (IsLess(c)) {
                tokenType = (sbyte)TokenLess;
            } else {
                tokenType = (sbyte)TokenSymbol;
            }
            if ( (tokenType != currentType )) {
                addToken();
                currentType = (sbyte)tokenType;
            }
            currentToken = SliceBuiltins.Append(currentToken, c);
        }
        addToken();
        return tokens;
    }

    public static void TokenizeTest()
    {
        var sqlStatement = new List<string> {@"Select * from table1 where field1 > 10;", @"(Select * from table1 where field1 > 10)", @"from table1 |> select *"};
        foreach (var statement in sqlStatement) {
            Console.WriteLine();
            Formatter.Printf(@"Tokenize:");
            Console.WriteLine(statement);
            var tokens1 = Tokenize(statement);
            foreach (var token in tokens1) {
                Formatter.Printf(DumpTokenString(token));
            }
        }
    }

}
}
namespace ast {

public class Api {

    public class LogicalExpr {
        public  lexer.Api.Token Value;
        public  ushort Left;
        public  ushort Right;
        public  List<LogicalExpr> Expressions;
    };

    public class Select {
        public  List<lexer.Api.Token> Fields;
        public  lexer.Api.Token ResultTableExpr;
    };

    public class Where {
        public  LogicalExpr Expr;
        public  lexer.Api.Token ResultTableExpr;
    };

    public class From {
        public  List<lexer.Api.Token> TableExpr;
        public  lexer.Api.Token ResultTableExpr;
    };

    public class Statement {
        public  sbyte Type;
        public  From FromF;
        public  Where WhereF;
        public  Select SelectF;
    };

    public class Visitor {
        public  Func<object, From,  object> PreVisitFrom;
        public  Func<object, From,  object> PostVisitFrom;
        public  Func<object, Where,  object> PreVisitWhere;
        public  Func<object, Where,  object> PostVisitWhere;
        public  Func<object, Select,  object> PreVisitSelect;
        public  Func<object, Select,  object> PostVisitSelect;
        public  Func<object, LogicalExpr,  object> PreVisitLogicalExpr;
        public  Func<object, LogicalExpr,  object> PostVisitLogicalExpr;
    };

    public const int StatementTypeFrom = 1;
    public const int StatementTypeWhere = 2;
    public const int StatementTypeSelect = 3;

    //using AST = List<Statement>;

    public static object WalkFrom(From expr, object state, Visitor visitor)
    {
        state = (object)visitor.PreVisitFrom(state, expr);
        state = (object)visitor.PostVisitFrom(state, expr);
        return (object)state;
    }

    public static object WalkWhere(Where where, object state, Visitor visitor)
    {
        state = (object)visitor.PreVisitWhere(state, where);
        state = (object)walkLogicalExpr(where.Expr, state, visitor);
        state = (object)visitor.PostVisitWhere(state, where);
        return (object)state;
    }

    public static object WalkSelect(Select expr, object state, Visitor visitor)
    {
        state = (object)visitor.PreVisitSelect(state, expr);
        state = (object)visitor.PostVisitSelect(state, expr);
        return (object)state;
    }

    public static object walkLogicalExpr(LogicalExpr expr, object state, Visitor visitor)
    {
        state = (object)visitor.PreVisitLogicalExpr(state, expr);
        if ( ( (expr.Left != 0 ) ||  (expr.Right != 0 ) )) {
            state = (object)walkLogicalExpr(expr.Expressions[0], state, visitor);
            state = (object)walkLogicalExpr(expr.Expressions[1], state, visitor);
        }
        state = (object)visitor.PostVisitLogicalExpr(state, expr);
        return (object)state;
    }

}
}
namespace parser {
using AST = List<ast.Api.Statement>;
public class Api {

    public class Node {
        public  sbyte Type;
        public  lexer.Api.Token Tok;
        public  List<short> Children;
        public  List<Node> Nodes;
    };

    public static sbyte precedence(List<sbyte> op)
    {
        if ( ( (op[0] == '&' ) &&  (op[1] == '&' ) )) {
            return (sbyte)1;
        }
        if ( ( (op[0] == '|' ) &&  (op[1] == '|' ) )) {
            return (sbyte)1;
        }
        if ( (op[0] == '>' )) {
            return (sbyte)2;
        }
        if ( (op[0] == '<' )) {
            return (sbyte)2;
        }
        if ( ( (op[0] == '>' ) &&  (op[1] == '=' ) )) {
            return (sbyte)2;
        }
        if ( ( (op[0] == '<' ) &&  (op[1] == '=' ) )) {
            return (sbyte)2;
        }
        if ( ( (op[0] == '=' ) &&  (op[1] == '=' ) )) {
            return (sbyte)2;
        }
        if ( ( (op[0] == '!' ) &&  (op[1] == '=' ) )) {
            return (sbyte)2;
        }
        return (sbyte)(-1);
    }

    public static sbyte associativity(List<sbyte> op)
    {
        if ( ( (op[0] == '&' ) &&  (op[1] == '&' ) )) {
            return (sbyte)'L';
        }
        if ( ( (op[0] == '|' ) &&  (op[1] == '|' ) )) {
            return (sbyte)'L';
        }
        if ( (op[0] == '>' )) {
            return (sbyte)'L';
        }
        if ( (op[0] == '<' )) {
            return (sbyte)'L';
        }
        if ( ( (op[0] == '>' ) &&  (op[1] == '=' ) )) {
            return (sbyte)'L';
        }
        if ( ( (op[0] == '<' ) &&  (op[1] == '=' ) )) {
            return (sbyte)'L';
        }
        if ( ( (op[0] == '=' ) &&  (op[1] == '=' ) )) {
            return (sbyte)'L';
        }
        if ( ( (op[0] == '!' ) &&  (op[1] == '=' ) )) {
            return (sbyte)'L';
        }
        return (sbyte)'L';
    }

    public static (ast.Api.LogicalExpr,int) ParseExpression(List<lexer.Api.Token> tokens)
    {
        var (expr, index) = parseExpression(tokens, 0);
        return (expr, index);
    }

    public static (ast.Api.LogicalExpr,int) parseExpression(List<lexer.Api.Token> tokens, sbyte minPrecedence)
    {
        var lhs = new ast.Api.LogicalExpr {Value= tokens[0]};
        var i = 1;
        for (; (i < SliceBuiltins.Length(tokens) );) {
            var token = tokens[i];
            var tokenPrecedence = (sbyte)precedence(token.Representation);
            if ( ( (tokenPrecedence == (-1) ) ||  (tokenPrecedence < minPrecedence ) )) {
                break;
            }
            var nextPrecedence = (sbyte)tokenPrecedence;
            if ( (associativity(token.Representation) == 'L' )) {
                nextPrecedence += (sbyte)1;
            }
            var (rhsExpr, nextPos) = parseExpression(tokens[ (i + 1 )..], nextPrecedence);
            var rhsIndex = (ushort)( (SliceBuiltins.Length(lhs.Expressions) + 1 ));
            lhs = new ast.Api.LogicalExpr {Value= token, Left= 0, Right= rhsIndex, Expressions= new List<ast.Api.LogicalExpr>{lhs, rhsExpr}};
            i +=  (nextPos + 1 );
        }
        return (lhs, i);
    }

    public static (ast.Api.From,List<lexer.Api.Token>) parseFrom(List<lexer.Api.Token> tokens, lexer.Api.Token lhs)
    {
        var from = new ast.Api.From {ResultTableExpr= lhs, TableExpr = new List<lexer.Api.Token>()};
        for (;;) {
            lexer.Api.Token token = new lexer.Api.Token{
                Representation = new List<sbyte>()
            };
            (token, tokens) = lexer.Api.GetNextToken(tokens);
            from.TableExpr = SliceBuiltins.Append(from.TableExpr, token);
            if (lexer.Api.IsSemicolon(token.Representation[0])) {
                break;
            }
        }
        return (from, tokens);
    }

    public static (ast.Api.Where,List<lexer.Api.Token>) parseWhere(List<lexer.Api.Token> tokens, lexer.Api.Token lhs)
    {
        var (expr, i) = ParseExpression(tokens);
        tokens = tokens[..];
        for (;;) {
            lexer.Api.Token token = default;
            (token, tokens) = lexer.Api.GetNextToken(tokens);
            if (lexer.Api.IsSemicolon(token.Representation[0])) {
                break;
            }
        }
        return (new ast.Api.Where {Expr= expr, ResultTableExpr= lhs}, tokens);
    }

    public static (ast.Api.Select,List<lexer.Api.Token>) parseSelect(List<lexer.Api.Token> tokens, lexer.Api.Token lhs)
    {
        var project = new ast.Api.Select {ResultTableExpr= lhs,
            Fields = new List<lexer.Api.Token>()
        };
        for (;;) {
            lexer.Api.Token token = new lexer.Api.Token{
                Representation = new List<sbyte>()
            };
            (token, tokens) = lexer.Api.GetNextToken(tokens);
            project.Fields = SliceBuiltins.Append(project.Fields, token);
            if (lexer.Api.IsSemicolon(token.Representation[0])) {
                break;
            }
        }
        return (project, tokens);
    }

    public static (AST,sbyte) Parse(string text)
    {
        AST resultAst = new AST();
        var tokens = lexer.Api.GetTokens(lexer.Api.StringToToken(text));
        lexer.Api.DumpTokens(tokens);
        for (; (SliceBuiltins.Length(tokens) > 0 );) {
            lexer.Api.Token token = default;
            (token, tokens) = lexer.Api.GetNextToken(tokens);
            if ((!lexer.Api.IsAlpha(token.Representation[0]))) {
                return (new AST {}, (-1));
            }
            var lhs = token;
            (token, tokens) = lexer.Api.GetNextToken(tokens);
            if ((!lexer.Api.IsEqual(token.Representation[0]))) {
                return (new AST {}, (-1));
            }
            (token, tokens) = lexer.Api.GetNextToken(tokens);
            if ( ( ((!lexer.Api.IsFrom(token)) && (!lexer.Api.IsWhere(token)) ) && (!lexer.Api.IsSelect(token)) )) {
                return (new AST {}, (-1));
            }
            if (lexer.Api.IsFrom(token)) {
                ast.Api.From from = new ast.Api.From();
                (from, tokens) = parseFrom(tokens, lhs);
                resultAst = SliceBuiltins.Append(resultAst, new ast.Api.Statement {Type= ast.Api.StatementTypeFrom, FromF= from});
                continue;
            }
            if (lexer.Api.IsWhere(token)) {
                ast.Api.Where where = new ast.Api.Where();
                (where, tokens) = parseWhere(tokens, lhs);
                resultAst = SliceBuiltins.Append(resultAst, new ast.Api.Statement {Type= ast.Api.StatementTypeWhere, WhereF= where});
                continue;
            }
            if (lexer.Api.IsSelect(token)) {
                ast.Api.Select project = new ast.Api.Select();

                (project, tokens) = parseSelect(tokens, lhs);
                resultAst = SliceBuiltins.Append(resultAst, new ast.Api.Statement {Type= ast.Api.StatementTypeSelect, SelectF= project});
                (token, tokens) = lexer.Api.GetNextToken(tokens);
                continue;
            }
        }
        return (resultAst, 0);
    }

}
}
namespace MainClass {

public class Api {

    public class State {
        public  sbyte depth;
    };

    public static void Main()
    {
        var visitor = new ast.Api.Visitor {PreVisitFrom= (object state, ast.Api.From expr)=>{
                {
                    var newState = (State)state;
                    newState.depth++;
                    return newState;
                }
            }, PostVisitFrom= (object state, ast.Api.From from)=>{
                {
                    var newState = (State)state;
                    Console.WriteLine(@"From:");
                    string result = default;
                    string indent = default;
                    for (var i = 0; (i < (int)(newState.depth) ); i++)
                    {
                        indent += (string)@"  ";
                    }
                    result += (string)indent;
                    result += (string)lexer.Api.DumpTokenString(from.ResultTableExpr);
                    result += (string)indent;
                    result += (string)lexer.Api.DumpTokenString(from.TableExpr[0]);
                    Formatter.Printf(result);
                    newState.depth--;
                    return newState;
                }
            }, PreVisitWhere= (object state, ast.Api.Where where)=>{
                {
                    var newState = (State)state;
                    newState.depth++;
                    Console.WriteLine(@"Where:");
                    string result = default;
                    string indent = default;
                    for (var i = 0; (i < (int)(newState.depth) ); i++)
                    {
                        indent += (string)@"  ";
                    }
                    result += (string)indent;
                    result += (string)lexer.Api.DumpTokenString(where.ResultTableExpr);
                    result += (string)indent;
                    result += (string)lexer.Api.DumpTokenString(where.ResultTableExpr);
                    Formatter.Printf(result);
                    newState.depth--;
                    return newState;
                }
            }, PostVisitWhere= (object state, ast.Api.Where expr)=>{
                {
                    var newState = (State)state;
                    return newState;
                }
            }, PreVisitSelect= (object state, ast.Api.Select project)=>{
                {
                    var newState = (State)state;
                    newState.depth++;
                    Console.WriteLine(@"Select:");
                    string result = default;
                    string indent = default;
                    for (var i = 0; (i < (int)(newState.depth) ); i++)
                    {
                        indent += (string)@"  ";
                    }
                    result += (string)indent;
                    result += (string)lexer.Api.DumpTokenString(project.ResultTableExpr);
                    result += (string)indent;
                    result += (string)lexer.Api.DumpTokenString(project.Fields[0]);
                    Formatter.Printf(result);
                    newState.depth--;
                    return newState;
                }
            }, PostVisitSelect= (object state, ast.Api.Select expr)=>{
                {
                    var newState = (State)state;
                    newState.depth--;
                    return newState;
                }
            }, PreVisitLogicalExpr= (object state, ast.Api.LogicalExpr expr)=>{
                {
                    var newState = (State)state;
                    newState.depth++;
                    string result = default;
                    string indent = default;
                    for (var i = 0; (i < (int)(newState.depth) ); i++)
                    {
                        indent += (string)@"  ";
                    }
                    result += (string)indent;
                    result += (string)lexer.Api.DumpTokensString(new List<lexer.Api.Token>{expr.Value});
                    Formatter.Printf(result);
                    return newState;
                }
            }, PostVisitLogicalExpr= (object state, ast.Api.LogicalExpr expr)=>{
                {
                    var newState = (State)state;
                    newState.depth--;
                    return newState;
                }
            }
        };
        var (astTree, err) = parser.Api.Parse(@"
 t1 = from table1;
 t2 = where t1.field1 > 10 && t1.field2 < 20;
 t3 = select t2.field1;
");
        if ( (err != 0 )) {
            Console.WriteLine(@"Error parsing query");
        }
        object state = default;
        state = new State {depth= 0};
        foreach (var statement in astTree) {
            Console.WriteLine(statement.Type);
            switch (statement.Type) {
            case (sbyte)ast.Api.StatementTypeFrom:
                state = (object)ast.Api.WalkFrom(statement.FromF, state, visitor);
                break;
            case (sbyte)ast.Api.StatementTypeWhere:
                state = (object)ast.Api.WalkWhere(statement.WhereF, state, visitor);
                break;
            case (sbyte)ast.Api.StatementTypeSelect:
                state = (object)ast.Api.WalkSelect(statement.SelectF, state, visitor);
                break;
            }
        }
        //lexer.Api.TokenizeTest();
    }

}
}
