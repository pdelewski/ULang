package types

import "fmt"

// Constants with explicit values (from substrait)
const (
	ExprLiteral  ExprKind = 0
	ExprColumn   ExprKind = 1
	ExprFunction ExprKind = 2
)

const (
	RelOpScan    RelNodeKind = 0
	RelOpFilter  RelNodeKind = 1
	RelOpProject RelNodeKind = 2
)

// Type aliases (from substrait)
type ExprKind int
type RelNodeKind int

// Struct with int32, int64 types (from iceberg)
type DataRecord struct {
	ID          int32
	Size        int64
	Count       int32
	SequenceNum int64
}

// Struct with bool field (from iceberg Field)
type Field struct {
	ID       int32
	Name     string
	Required bool
}

// Struct with nested struct field - not slice (from iceberg ManifestEntry)
type ColumnStats struct {
	NullCount int64
}

type DataFile struct {
	FilePath    string
	RecordCount int64
	Stats       ColumnStats
}

type ManifestEntry struct {
	Status    int32
	DataFileF DataFile
}

// Struct with custom type fields (from substrait)
type ExprNode struct {
	Kind     ExprKind
	Children []int
	ValueIdx int
}

type RelNode struct {
	Kind     RelNodeKind
	InputIdx int
	ExprIdxs []int
}

// Struct with slices of structs (from substrait)
type Plan struct {
	RelNodes []RelNode
	Exprs    []ExprNode
	Literals []string
	Root     int
}

// Helper function in package (from substrait pattern)
func AddLiteralToPlan(plan Plan, value string) (Plan, int) {
	plan.Literals = append(plan.Literals, value)
	return plan, len(plan.Literals) - 1
}

// Function with struct parameter (from iceberg catalog)
func LoadData(record DataRecord) {
	fmt.Println("Loading data")
	fmt.Println(record.ID)
}
