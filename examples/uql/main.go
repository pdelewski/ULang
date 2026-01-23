package main

import (
	"fmt"
	"uql/emitter"
	"uql/lexer"
	"uql/parser"
)

func main() {
	astTree, err := parser.Parse(`
 t1 = from table1;
 t2 = from table2;
 t3 = join t1 t2 on t1.id == t2.id;
 t4 = where t3.field1 > 10 && t3.field2 < 20;
 t5 = orderby t4 t4.field1 desc t4.field2 asc;
 t6 = limit t5 100;
 t7 = groupby t4 t4.category count t4.id sum t4.amount;
 t8 = select t6.field1;
`)

	if err != 0 {
		fmt.Println("Error parsing query")
	}

	result := emitter.EmitUql(astTree)
	fmt.Print(result)

	lexer.TokenizeTest()
}
