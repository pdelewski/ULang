package catalog

// ManifestFile represents an Iceberg manifest file.
type ManifestFile struct {
	ManifestPath       string          `json:"manifest_path"`        // Path to the manifest file
	ManifestLength     int64           `json:"manifest_length"`      // File size in bytes
	PartitionSpecID    int32           `json:"partition_spec_id"`    // ID of the partition spec used
	Content            int32           `json:"content"`              // 0 for data files, 1 for delete files
	SequenceNumber     int64           `json:"sequence_number"`      // Tracks table evolution
	MinSequenceNumber  int64           `json:"min_sequence_number"`  // Minimum sequence number of referenced files
	SnapshotID         int64           `json:"snapshot_id"`          // Snapshot that added this manifest
	AddedFilesCount    int32           `json:"added_files_count"`    // Number of added data files
	ExistingFilesCount int32           `json:"existing_files_count"` // Number of existing files
	DeletedFilesCount  int32           `json:"deleted_files_count"`  // Number of deleted files
	CreatedAt          int32           `json:"created_at"`           // Manifest creation timestamp
	Partitions         []Partition     `json:"partitions"`           // Partition metadata
	Entries            []ManifestEntry `json:"entries"`              // List of data file entries
}

// Partition represents partition-related metadata.
type Partition struct {
	Fields []interface{} `json:"fields"` // Partition field values
}

// ManifestEntry represents a single data file entry in the manifest.
type ManifestEntry struct {
	Status     int32    `json:"status"`      // 0 = existing, 1 = added, 2 = deleted
	SnapshotID int64    `json:"snapshot_id"` // Snapshot that added this file
	DataFileF  DataFile `json:"data_file"`   // Associated data file
}

// DataFile represents a data file in the manifest.
type DataFile struct {
	FilePath        string        `json:"file_path"`       // Path to the data file
	FileFormat      string        `json:"file_format"`     // Parquet, ORC, Avro
	Partition       []interface{} `json:"partition"`       // Partition values
	RecordCount     int64         `json:"record_count"`    // Number of records
	FileSizeInBytes int64         `json:"file_size_bytes"` // File size in bytes
	ColumnStatsF    []ColumnStats `json:"column_stats"`    // Statistics per column
}

// ColumnStats represents statistics for a single column.
type ColumnStats struct {
	MinValue  interface{} `json:"min_value"`  // Minimum value
	MaxValue  interface{} `json:"max_value"`  // Maximum value
	NullCount int64       `json:"null_count"` // Number of null values
}

// ManifestList represents an Iceberg manifest list file
type ManifestList struct {
	Version       int32          `json:"version"`
	ManifestFiles []ManifestFile `json:"manifest_files"`
	CreatedAt     int64          `json:"created_at"`
}

// MetadataFile represents the Iceberg metadata.json file
type MetadataFile struct {
	FormatVersion     int             `json:"format-version"`
	TableUUID         string          `json:"table-uuid"`
	Location          string          `json:"location"`
	LastUpdatedMs     int64           `json:"last-updated-ms"`
	CurrentSnapshotID int64           `json:"current-snapshot-id,omitempty"`
	Snapshots         []ManifestList  `json:"snapshots,omitempty"`
	PartitionSpecs    []PartitionSpec `json:"partition-specs"`
	Schemas           []Schema        `json:"schemas"`
}

// Schema defines the schema structure in Iceberg
type Schema struct {
	SchemaID int32   `json:"schema-id"`
	Fields   []Field `json:"fields"`
}

// Field represents a column in the schema
type Field struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// PartitionSpec defines partitioning rules
type PartitionSpec struct {
	SpecID int32            `json:"spec-id"`
	Fields []PartitionField `json:"fields"`
}

// PartitionField defines a partition field
type PartitionField struct {
	SourceID  int32  `json:"source-id"`
	Transform string `json:"transform"`
	Name      string `json:"name"`
}
