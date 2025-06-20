package catalog

import "fmt"

type Catalog struct {
}

func LoadTable(catalog Catalog, table string) {
	fmt.Println("Loading table")
}
