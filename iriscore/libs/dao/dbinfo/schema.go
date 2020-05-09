package dbinfo

import "goweb/iriscore/libs/dao/dbnulltype"

type TDBTab struct {
	TableName string                 `db:"table_name" json:"table_name"`
	RowCount  dbnulltype.NullFloat64 `db:"table_rows" json:"table_rows"`
	TableMb   dbnulltype.NullFloat64 `db:"used_mb" json:"used_mb"`
}

// column
type TDBCol struct {
	//ColumnID     int    `db:"column_id" json:"column_id"`
	ColumnName   string `db:"column_name" json:"column_name"`
	Comment      string `db:"column_comment" json:"column_comment"`
	DataType     string `db:"column_type" json:"column_type"`
	ColumnLength int    `db:"column_length" json:"column_length"`
	Precision    int    `db:"column_precision" json:"column_precision"`
	IsNull       string `db:"is_nullable" json:"is_nullable"`
	IsIdentity   string `db:"is_identity" json:"is_identity"`
	IsPrimary    string `db:"is_primary" json:"is_primary"`
	DefaultValue string `db:"column_default" json:"column_default"`
}
