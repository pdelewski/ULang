package main

import uqlparser "ULang/uql/parser"

func main() {

	uqlparser.Parse(`
 t1 = from table1
 t2 = where t1.field1 > 10
 t3 = select t2.field1
`)

}
