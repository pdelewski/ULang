package main

import (
	"iceberg/catalog"
)

func main() {
	c := catalog.Catalog{}
	catalog.LoadTable(c, "table")
}
