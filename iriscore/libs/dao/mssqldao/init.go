package mssqldao

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	proto "goweb/iriscore/libs/dao/dbinfo"
	log "goweb/iriscore/libs/logrus"

	dao "goweb/iriscore/libs/dao"
	dbtype "goweb/iriscore/libs/dao/dbnulltype"
	daotmout "goweb/iriscore/libs/dao/dbtimeout"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/go-gorp/gorp"
)

const (
	DB_TYPE         = "mssql"
	DB_VERSION      = "2012"
	COL_SELECT_SPEC = `
SELECT  
        col.name AS COLUMN_NAME ,
        t.name+'('+CONVERT(VARCHAR(100), col.length)+')' AS COLUMN_TYPE ,
        CASE WHEN COLUMNPROPERTY(col.id, col.name, 'IsIdentity') = 1 THEN 'YES'
             ELSE 'NO'
        END AS is_identity ,
        CASE WHEN EXISTS ( SELECT   1
                           FROM     dbo.sysindexes si
                                    INNER JOIN dbo.sysindexkeys sik ON si.id = sik.id
                                                              AND si.indid = sik.indid
                                    INNER JOIN dbo.syscolumns sc ON sc.id = sik.id
                                                              AND sc.colid = sik.colid
                                    INNER JOIN dbo.sysobjects so ON so.name = si.name
                                                              AND so.xtype = 'PK'
                           WHERE    sc.id = col.id
                                    AND sc.colid = col.colid ) THEN 'YES'
             ELSE 'NO'
        END AS IS_PRIMARY ,
        CASE WHEN col.isnullable = 1 THEN 'YES'
             ELSE 'NO'
        END AS IS_NULLABLE ,
convert(varchar(8000),ISNULL(ep.[value], '') )AS COLUMN_COMMENT ,
        case when comm.text is null then 'null' else  comm.text end AS COLUMN_DEFAULT

FROM    dbo.syscolumns col
        LEFT  JOIN dbo.systypes t ON col.xtype = t.xusertype
        inner JOIN dbo.sysobjects obj ON col.id = obj.id
                                         AND obj.xtype = 'U'
                                         AND obj.status >= 0
        LEFT  JOIN dbo.syscomments comm ON col.cdefault = comm.id
        LEFT  JOIN sys.extended_properties ep ON col.id = ep.major_id
                                                      AND col.colid = ep.minor_id
                                                      AND ep.name = 'MS_Description'
        LEFT  JOIN sys.extended_properties epTwo ON obj.id = epTwo.major_id
                                                         AND epTwo.minor_id = 0
                                                         AND epTwo.name = 'MS_Description'
WHERE   obj.name = ? ;
`

	TAB_SELECT_SPEC = `
declare @db varchar(100)
set @db = ? 
IF NOT EXISTS(SELECT * FROM sys.databases WHERE Name=@db)
BEGIN
PRINT @db + 'database not exist'
END
declare @sql VARCHAR(8000)
set @sql='begin try 

SELECT 
TABLE_NAME,TABLE_ROWS,
round(convert(numeric(25,10),RESERVED)/convert(numeric(25,10),1024),4) AS used_mb
FROM

  (SELECT
    a3.name AS [schemaname],
    a2.name AS [Table_name],
    a1.rows as Table_rows,
    (a1.reserved + ISNULL(a4.reserved,0))* 8 AS Reserved, 
    a1.data * 8 AS Data,
    (CASE WHEN (a1.used + ISNULL(a4.used,0)) > a1.data 
    THEN (a1.used + ISNULL(a4.used,0)) - a1.data ELSE 0 END) * 8 AS Index_size,
    (CASE WHEN (a1.reserved + ISNULL(a4.reserved,0)) > a1.used 
    THEN (a1.reserved + ISNULL(a4.reserved,0)) - a1.used ELSE 0 END) * 8 AS Unused
    FROM
        (SELECT ps.object_id,
          SUM (
          CASE
          WHEN (ps.index_id < 2) THEN row_count ELSE 0
          END
          ) AS [rows],
          SUM (ps.reserved_page_count) AS reserved,
          SUM (
          CASE
          WHEN (ps.index_id < 2) THEN 
          (ps.in_row_data_page_count +
          ps.lob_used_page_count + ps.row_overflow_used_page_count)
          ELSE (ps.lob_used_page_count + ps.row_overflow_used_page_count)
          END
          ) AS data,
          SUM (ps.used_page_count) AS used
          FROM ' + @db + '.sys.dm_db_partition_stats ps
          GROUP BY ps.object_id) AS a1
    LEFT OUTER JOIN 
        (SELECT 
        it.parent_id,
        SUM(ps.reserved_page_count) AS reserved,
        SUM(ps.used_page_count) AS used
        FROM ' + @db + '.sys.dm_db_partition_stats ps
        INNER JOIN ' + @db +'.sys.internal_tables it ON (it.object_id = ps.object_id)
        WHERE it.internal_type IN (202,204)
        GROUP BY it.parent_id) AS a4 ON (a4.parent_id = a1.object_id)
        INNER JOIN ' + @db +'.sys.all_objects a2  ON ( a1.object_id = a2.object_id ) 
        INNER JOIN ' + @db +'.sys.schemas a3 ON (a2.schema_id = a3.schema_id)
        WHERE a2.type <> N''S'' and a2.type <> N''IT''
        ) Basic
        ORDER BY TABLE_NAME
    end try 
    begin catch 
    select 
               ERROR_NUMBER() as table_rows
       ,       ERROR_SEVERITY() as used_mb
       ,       ERROR_MESSAGE() as table_name
    end catch'
EXEC(@sql)
`
)

var (
	ERRNilDB  = errors.New("nilmssql conn")
	ERRNilObj = errors.New("nil obj")
	ERRDBConn = errors.New("conn db err")

	ZeroId = 0
)

var (
	DBMSMgr *MsSqlConn
)

func GenMssqlConn() dao.DbConn {
	return &MssqlConn{}
}

func init() {
	dao.RegisterFactory("oracle", GenMssqlConn)
	dao.RegisterFactory("Oracle", GenMssqlConn)
}

func InitDbConn(dbaddr string) error {
	DBMSMgr = &MsSqlConn{}
	return DBMSMgr.Init(dbaddr)
}

func CloseDbConn() {
	DBMSMgr.Close()
}

func GetDbConn() *MsSqlConn {
	return DBMSMgr
}

type MsSqlConn struct {
	DBObj    *gorp.DbMap
	DBConn   *sql.DB
	dataaddr string //connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;connection timeout=30", *server, *user, *password, *port, *database)
}

func (c *MsSqlConn) Init(dataaddr string) error {
	var err error
	c.dataaddr = dataaddr
	c.DBConn, err = sql.Open(DB_TYPE, dataaddr)
	if err != nil {
		return err
	}
	c.DBObj = &gorp.DbMap{Db: c.DBConn, Dialect: gorp.SqlServerDialect{DB_VERSION}}
	// grop debug log
	//c.DBObj.TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))
	return nil
}

// result must be an  []
func (c *MsSqlConn) SelectWithOrm(result interface{}, sqlstmt string) error {
	_, err := c.DBObj.Select(result, sqlstmt)
	return err
}

func (c *MsSqlConn) Killsql(db, sql string) (bool, error) {
	return false, nil
}

func (c *MsSqlConn) Execute(sqlstmt string) (int, error) {
	if rets, err := c.DBObj.Exec(sqlstmt); err != nil {
		return 0, err
	} else {
		rows, _ := rets.RowsAffected()
		return int(rows), nil
	}
}

func (c *MsSqlConn) BatchExecute(sqls []string) error {

	tx, err := c.DBObj.Begin()
	if err != nil {
		return err
	}

	for _, sql := range sqls {
		_, err = tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return err
}

type columnType struct {
	Ty    int
	valid bool
}

func (c *MsSqlConn) Select(sqlstmt string) ([][]string, []string, []int, error) {
	var ret [][]string
	//rows, err := c.DBConn.Query(sqlstmt)
	ctx, cancel := context.WithTimeout(context.Background(), daotmout.CONNECTION_TIMEOUT)
	defer cancel()

	rows, err := c.DBConn.QueryContext(ctx, sqlstmt)
	if err != nil {
		return nil, nil, nil, err
	}

	defer rows.Close()
	//select {
	//case <-ctx.Done():
	//	return nil, nil, errors.New(fmt.Sprintf("查询链接超时: %s", ctx.Err().Error()))
	//}

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, nil, err
	}
	if cols == nil {
		return nil, nil, nil, fmt.Errorf("nil columns")
	}
	coltys, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, nil, err
	}
	var colinfo []*columnType
	for _, cv := range coltys {
		log.Debugf("!!!!!!!!!!!!!!column:%s  golang_type:%s db_typ:%s", cv.Name(), cv.ScanType().Name(), cv.DatabaseTypeName())
		coltemp := &columnType{}
		coltemp.Ty = dbtype.ValueType_TUnKnown
		colinfo = append(colinfo, coltemp)
	}
	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(interface{})
	}
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return nil, nil, nil, err
		}
		var strv []string
		for i, v := range vals {
			str, ty := TostringType(v.(*interface{}), strings.ToLower(coltys[i].DatabaseTypeName()))
			strv = append(strv, str)
			if !colinfo[i].valid {
				colinfo[i].Ty = ty
				colinfo[i].valid = true
			}
		}
		ret = append(ret, strv)
	}
	if rows.Err() != nil {
		return nil, nil, nil, rows.Err()
	}
	var colty []int
	for _, co := range colinfo {
		colty = append(colty, co.Ty)
	}
	return ret, cols, colty, nil
}
func (c *MsSqlConn) SelectValues(sqlstmt string) ([][]string, []string, error) {
	var ret [][]string
	//rows, err := c.DBConn.Query(sqlstmt)
	ctx, cancel := context.WithTimeout(context.Background(), dao.CONNECTION_TIMEOUT)
	defer cancel()

	rows, err := c.DBConn.QueryContext(ctx, sqlstmt)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()
	//select {
	//case <-ctx.Done():
	//	return nil, nil, errors.New(fmt.Sprintf("查询链接超时: %s", ctx.Err().Error()))
	//}

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	if cols == nil {
		return nil, nil, fmt.Errorf("nil columns")
	}
	coltys, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}
	for _, cv := range coltys {
		log.Debugf("!!!!!!!!!!!!!!column:%s  golang_type:%s db_typ:%s", cv.Name(), cv.ScanType().Name(), cv.DatabaseTypeName())
	}
	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(interface{})
	}
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return nil, nil, err
		}
		var strv []string
		for i, v := range vals {
			strv = append(strv, Tostring(v.(*interface{}), strings.ToLower(coltys[i].DatabaseTypeName())))
		}
		ret = append(ret, strv)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}
	return ret, cols, nil
}

//
func (c *MsSqlConn) GetDbConn() *sql.DB {
	return c.DBConn
}

func (c *MsSqlConn) Ping() error {
	return c.DBConn.Ping()
}

//
func (c *MsSqlConn) GetDbObj() *gorp.DbMap {
	return c.DBObj
}

//
func (c *MsSqlConn) Close() {
	c.DBConn.Close()
}

func (c *MsSqlConn) Makedatasource(ipport, dbname string, user, pass string) string {
	addrs := strings.Split(ipport, ":")
	if len(addrs) != 2 {
		return ""
	}
	return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;connection timeout=3600", addrs[0], user, pass, addrs[1], dbname)
}

func (t *MsSqlConn) SelectTableForDb(dbname string) ([]proto.TDBTab, error) {

	begin := time.Now()
	var tabs []proto.TDBTab
	if _, err := t.DBObj.Select(&tabs, TAB_SELECT_SPEC, dbname); err != nil {
		log.Errorf("MSSQL::Get TAblelisti for %s::err:%v", dbname, err)
		//log.Errorf("MSSQL::Get TAblelisti for %s %s", TAB_SELECT_SPEC, dbname)
		return nil, err
	}
	log.Debugf("get sqlserver tabinfo cost:%s", time.Now().Sub(begin).String())
	return tabs, nil
}

func (t *MsSqlConn) SelectColumnForTable(dbname, tabname string) ([]proto.TDBCol, error) {

	begin := time.Now()
	var cols []proto.TDBCol
	if _, err := t.DBObj.Select(&cols, COL_SELECT_SPEC, tabname); err != nil {
		return nil, err
	}
	log.Debugf("get sqlserver columninfo cost:%s", time.Now().Sub(begin).String())
	return cols, nil
}

func binaryDataConv(data []byte, typename string) string {
	typename = strings.ToLower(typename)
	switch {
	case strings.HasPrefix(typename, "money"):
		fallthrough
	case strings.HasPrefix(typename, "smallmoney"):
		return string(data)
	case strings.HasPrefix(typename, "numeric"):
		fallthrough
	case strings.HasPrefix(typename, "decimal"):
		return string(data)
	case strings.HasPrefix(typename, "blob"):
		return "(blob)"
	default:
		return "0x" + hex.EncodeToString(data)
	}

	return ""
}

func timeDataConv(v time.Time, typename string) string {
	typename = strings.ToLower(typename)
	switch {
	case strings.HasPrefix(typename, "datetime"):
		return fmt.Sprintf("%s", v.Format("2006-01-02 15:04:05.000"))

	case strings.HasPrefix(typename, "datetime2"):
		return fmt.Sprintf("%s", v.Format("2006-01-02 15:04:05.0000000"))

	case strings.HasPrefix(typename, "smalldatetime"):
		return fmt.Sprintf("%s", v.Format("2006-01-02 15:04"))

	case strings.HasPrefix(typename, "date"):
		return fmt.Sprintf("%s", v.Format("2006-01-02"))

	case strings.HasPrefix(typename, "time"):
		return fmt.Sprintf("%s", v.Format("15:04:05.0000000"))

	case strings.HasPrefix(typename, "datetimeoffset"):
		return fmt.Sprintf("%s", v.Format("2006-01-02 15:04:05.0000000 -07:00"))
	default:
		return fmt.Sprintf("%s", v.Format("2006-01-02 15:04:05.000"))
	}

	return ""
}

func Tostring(av interface{}, typeinfo string) string {
	tt, ok := av.(*interface{})
	if !ok {
		return ""
	}
	switch v := (*tt).(type) {
	case nil:
		return "NULL"
	case bool:
		if v {
			return "1"
		} else {
			return "0"
		}
	case []byte:
		return binaryDataConv(v, typeinfo)
	case *time.Time:
		return fmt.Sprintf("%s", (*v).Format("2006-01-02 15:04:05.000"))
	case byte:
		return "0x" + hex.EncodeToString([]byte{v})
	case string:
		return fmt.Sprintf("%s", v)
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return timeDataConv(v, typeinfo)
	case mssql.Decimal:
		return v.String()
	case mssql.UniqueIdentifier:
		return "0x" + hex.EncodeToString(v[:])

	default:
		return "NULL"
	}
	return "NULL"
}

func TostringType(av interface{}, typeinfo string) (string, int) {
	tt, ok := av.(*interface{})
	if !ok {
		return "", dbtype.ValueType_TUnKnown
	}
	switch v := (*tt).(type) {
	case nil:
		return "NULL", dbtype.ValueType_TString
	case bool:
		if v {
			return "1", dbtype.ValueType_TString
		} else {
			return "0", dbtype.ValueType_TString
		}
	case []byte:
		return binaryDataConv(v, typeinfo), dbtype.ValueType_TBinary
	case *time.Time:
		return fmt.Sprintf("%s", (*v).Format("2006-01-02 15:04:05.000")), dbtype.ValueType_TTime
	case byte:
		return "0x" + hex.EncodeToString([]byte{v}), dbtype.ValueType_TBinary
	case string:
		return fmt.Sprintf("%s", v), dbtype.ValueType_TString
	case int32:
		return strconv.Itoa(int(v)), dbtype.ValueType_TInt
	case int64:
		return strconv.FormatInt(v, 10), dbtype.ValueType_TInt
	case int:
		return strconv.Itoa(v), dbtype.ValueType_TInt
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64), dbtype.ValueType_TFloat
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), dbtype.ValueType_TFloat
	case time.Time:
		return timeDataConv(v, typeinfo), dbtype.ValueType_TTime

	default:
		return "NULL", dbtype.ValueType_TUnKnown
	}
	return "NULL", dbtype.ValueType_TUnKnown
}
