package models

//demo for sql
type Somestruct struct {
	Name      string `db:"login_name" json:"login_name"`
	Pwd       string `db:"pwd" json:"pwd"`
	Logintype int    `db:"type" json:"-"`
	IUD       string `db:"-" json:"login_type"`
	Dbtype    string `db:"-" json:"db_type"`
	Dbname    string `db:"-" json:"db_name"`
	Dbaddr    string `db:"-" json:"db_addr"`
	Idc       string `db:"region" json:"region"`
}

// demo for mongo
type DBA struct {
	Name string `bson:"name"`
	Role string `bson:"role"`
}
