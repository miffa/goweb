package oracledao

import (
	"DBMS_DBSEARCH/iriscore/proto"
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"DBMS_DBSEARCH/lib/dao"
	log "DBMS_DBSEARCH/lib/logrus"

	"github.com/go-gorp/gorp"
	_ "gopkg.in/rana/ora.v4" //oracle driver
	//_ "github.com/go-goracle/goracle"
	//_ "github.com/mattn/go-oci8"
)

const (
	DB_TYPE = "ora" // for ora.v4
	//DB_TYPE    = "goracle"  // for goracle
	DB_VERSION = "2012"

	DB_SCHEMA_CHAR = "::"

	TABLE_SELECT_SPEC = `select 
    a.table_name, a.num_rows as TABLE_ROWS ,round(b.BYTES/1024/1024,2) as used_mb from dba_tables  a
	inner join dba_segments b on a.table_name=b.segment_name 
	where a.owner=:1
	    and b.owner=:2`

	COLUMN_SELECT_SPEC = `select a.column_name , 
    '-' as is_identity, 
	'-' as IS_PRIMARY,
	a.data_type as COLUMN_TYPE, 
	a.nullable as IS_NULLABLE, 
	'-' as COLUMN_COMMENT,  
	a.data_default as column_default 
	from all_tab_columns a
    where 
	    a.table_name =:1 
	and 
	    a.owner=:2`
)

var (
	ERRNilDB  = errors.New("niloracle conn")
	ERRNilObj = errors.New("nil obj")
	ERRDBConn = errors.New("conn db err")

	ZeroId = 0
)

var (
	DBMSMgr *OracleConn
)

func InitDbConn(dbaddr string) error {
	DBMSMgr = &OracleConn{}
	return DBMSMgr.Init(dbaddr)
}

func CloseDbConn() {
	DBMSMgr.Close()
}

type OracleConn struct {
	DBObj    *gorp.DbMap
	DBConn   *sql.DB
	dataaddr string //
	dbname   string
}

func (c *OracleConn) Init(dataaddr string) error {
	var err error
	c.dataaddr = dataaddr
	c.DBConn, err = sql.Open(DB_TYPE, dataaddr)
	if err != nil {
		return err
	}
	c.DBObj = &gorp.DbMap{Db: c.DBConn, Dialect: gorp.OracleDialect{}}
	//c.DBObj.DynamicTableFor("haha", false)
	log.Debugf("connect %s ok.....", dataaddr)
	return nil
}

//
func (c *OracleConn) GetDbConn() *sql.DB {
	return c.DBConn
}

//
func (c *OracleConn) GetDbObj() *gorp.DbMap {
	return c.DBObj
}

//
func (c *OracleConn) Close() {
	c.DBConn.Close()
}
func (c *OracleConn) Execute(sqlstmt string) (int, error) {
	tx, err := c.DBConn.Begin()
	if err != nil {
		return 0, err
	}

	sessionsql := fmt.Sprintf("alter session set current_schema=%s", c.dbname)
	_, err = tx.Exec(sessionsql)
	if err != nil {
		log.Errorf("ORACLE::alter oracle session %s:%s err:%v", c.dbname, sessionsql, err)
		tx.Rollback()
		return 0, err
	}
	if rets, err := tx.Exec(sqlstmt); err != nil {
		tx.Rollback()
		return 0, err
	} else {
		rows, _ := rets.RowsAffected()
		tx.Commit()
		return int(rows), nil
	}
}

func (c *OracleConn) Ping() error {
	return c.DBConn.Ping()
}

func (c *OracleConn) SelectValues(sqlstmt string) ([][]interface{}, []string, error) {
	tx, err := c.DBConn.Begin()
	if err != nil {
		return nil, nil, err
	}

	sessionsql := fmt.Sprintf("alter session set current_schema=%s", c.dbname)
	_, err = tx.Exec(sessionsql)
	if err != nil {
		log.Errorf("ORACLE::alter oracle session %s:%s err:%v", c.dbname, sessionsql, err)
		tx.Rollback()
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), dao.CONNECTION_TIMEOUT)
	defer cancel()
	rows, err := tx.QueryContext(ctx, sqlstmt)
	if err != nil {
		log.Errorf("ORACLE::quert sql %s err:%v", sqlstmt, err)
		tx.Rollback()
		return nil, nil, err
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Errorf("ORACLE::quert sql %s err:%v", sqlstmt, err)
		return nil, nil, err
	}
	if cols == nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("nil columns")
	}
	coltys, err := rows.ColumnTypes()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	var ret [][]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		for i := 0; i < len(cols); i++ {
			vals[i] = new(interface{})
		}
		err = rows.Scan(vals...)
		if err != nil {
			tx.Rollback()
			return nil, nil, err
		}
		for i, v := range vals {
			vals[i] = Tostring(v, strings.ToLower(coltys[i].DatabaseTypeName()))
			log.Debugf("%s:%s:%s:%v:%T", cols[i], coltys[i].DatabaseTypeName(), vals[i], *(v.(*interface{})), *(v.(*interface{})))

		}
		ret = append(ret, vals)
		//log.Debugf("vals:%v", vals)
	}
	if rows.Err() != nil {
		tx.Rollback()
		log.Errorf("ORACLE::quert sql %s err:%v", sqlstmt, err)
		return nil, nil, rows.Err()
	}

	tx.Rollback()
	log.Debugf("ORACLE::quert sql %s ok len(%d)", sqlstmt, len(ret))
	return ret, cols, nil
}

func (c *OracleConn) BatchExecute(sqls []string) error {

	tx, err := c.DBObj.Begin()
	if err != nil {
		return err
	}

	sessionsql := fmt.Sprintf("alter session set current_schema=%s", c.dbname)
	_, err = tx.Exec(sessionsql)
	if err != nil {
		log.Errorf("ORACLE::alter oracle session %s:%s err:%v", c.dbname, sessionsql, err)
		tx.Rollback()
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

func (c *OracleConn) SelectWithOrm(result interface{}, sqlstmt string) error {
	tx, err := c.DBObj.Begin()
	if err != nil {
		return err
	}

	sessionsql := fmt.Sprintf("alter session set current_schema=%s", c.dbname)
	_, err = tx.Exec(sessionsql)
	if err != nil {
		log.Errorf("ORACLE::alter oracle session %s:%s err:%v", c.dbname, sessionsql, err)
		tx.Rollback()
		return err
	}
	_, err = tx.Select(result, sqlstmt)
	tx.Commit()
	return err
}

func (c *OracleConn) SelectStringValues(sqlstmt string) ([][]string, []string, error) {
	tx, err := c.DBConn.Begin()
	if err != nil {
		return nil, nil, err
	}
	sessionsql := fmt.Sprintf("alter session set current_schema=%s", c.dbname)
	_, err = tx.Exec(sessionsql)
	if err != nil {
		log.Errorf("ORACLE::alter oracle session %s:%s err:%v", c.dbname, sessionsql, err)
		tx.Rollback()
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), dao.CONNECTION_TIMEOUT)
	defer cancel()
	rows, err := tx.QueryContext(ctx, sqlstmt)
	if err != nil {
		log.Errorf("ORACLE::quert sql %s err:%v", sqlstmt, err)
		tx.Rollback()
		return nil, nil, err
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		tx.Rollback()
		log.Errorf("ORACLE::quert sql %s err:%v", sqlstmt, err)
		return nil, nil, err
	}
	if cols == nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("nil columns")
	}
	coltys, err := rows.ColumnTypes()
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	var ret [][]string
	vals := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		vals[i] = new(interface{})
	}

	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			tx.Rollback()
			return nil, nil, err
		}

		var stringvalues []string
		for i, v := range vals {
			stringvalues = append(stringvalues, Tostring(v, strings.ToLower(coltys[i].DatabaseTypeName())))
		}
		ret = append(ret, stringvalues)
	}
	if rows.Err() != nil {
		tx.Rollback()
		return nil, nil, rows.Err()
	}
	tx.Commit()
	return ret, cols, nil
}

func (c *OracleConn) Killsql(db, sql string) (bool, error) {
	return false, nil
}

//user/passw@host:port/sid
func (c *OracleConn) Makedatasource(ipport, dbname string, user, pass string) string {
	//c.dbname = dbname
	////return fmt.Sprintf("oracle://%s:%s@%s", user, pass, dbname) // for goracle
	//return fmt.Sprintf("%s/%s@%s/%s", user, pass, ipport, dbname)

	dbince := strings.Split(dbname, DB_SCHEMA_CHAR)
	c.dbname = dbince[len(dbince)-1]
	return fmt.Sprintf("%s/%s@%s/%s", user, pass, ipport, dbince[0])
}

func (c *OracleConn) SelectTableForDb(dbname string) ([]proto.TDBTab, error) {
	begin := time.Now()

	tx, err := c.DBObj.Begin()
	if err != nil {
		return nil, err
	}
	sessionsql := fmt.Sprintf("alter session set current_schema=%s", c.dbname)
	_, err = tx.Exec(sessionsql)
	if err != nil {
		log.Errorf("ORACLE::alter oracle session %s:%s err:%v", c.dbname, sessionsql, err)
		tx.Rollback()
		return nil, err
	}
	log.Debugf("ORACLE::alter oracle session %s:%s", c.dbname, sessionsql)

	var tabs []proto.TDBTab
	_, err = tx.Select(&tabs, TABLE_SELECT_SPEC, strings.ToUpper(c.dbname), strings.ToUpper(c.dbname))
	if err != nil {
		log.Errorf("ORACLE::quert sql %s err:%v", TABLE_SELECT_SPEC, err)
		tx.Rollback()
		return nil, err
	}

	log.Debugf("ORACLE::get tablelist:%s:%s ok  len(%d) cost:%s", TABLE_SELECT_SPEC, dbname, len(tabs), time.Now().Sub(begin).String())
	tx.Commit()
	return tabs, nil
}

func (t *OracleConn) SelectColumnForTable(dbname, tabname string) ([]proto.TDBCol, error) {

	begin := time.Now()

	tx, err := t.DBObj.Begin()
	if err != nil {
		return nil, err
	}
	sessionsql := fmt.Sprintf("alter session set current_schema=%s", t.dbname)
	_, err = tx.Exec(sessionsql)
	if err != nil {
		log.Errorf("ORACLE::alter oracle session %s:%s err:%v", t.dbname, sessionsql, err)
		tx.Rollback()
		return nil, err
	}
	log.Debugf("ORACLE::alter oracle session %s:%s err:%v", t.dbname, tabname, sessionsql)

	var cols []proto.TDBCol
	if _, err = tx.Select(&cols, COLUMN_SELECT_SPEC, strings.ToUpper(tabname), strings.ToUpper(t.dbname)); err != nil {
		tx.Rollback()
		return nil, err
	}
	log.Debugf("ORACLE::get tablelist:%s:%s:%s ok  len(%d) cost:%s", COLUMN_SELECT_SPEC, tabname, t.dbname, len(cols), time.Now().Sub(begin).String())
	tx.Commit()
	return cols, nil
}

func binaryDataConv(data []byte, typename string) string {
	typename = strings.ToLower(typename)
	switch {
	case strings.HasPrefix(typename, "raw"):
		return "(RAW)"
	case strings.HasPrefix(typename, "long raw"):
		return "(long raw)"
	case strings.HasPrefix(typename, "blob"):
		return "(blob)"
	default:
		return "0x" + hex.EncodeToString(data)
	}

	return ""
}

// todo: Confirm the display mode of different time formats
func timeDataConv(v time.Time, typename string) string {
	typename = strings.ToLower(typename)
	switch {
	case strings.HasPrefix(typename, "timestamp with time zone"):
		return fmt.Sprintf("%s", v.Format("2006/01/02 15:04:05.000"))

	case strings.HasPrefix(typename, "timestamp with local"):
		return fmt.Sprintf("%s", v.Format("2006/01/02 15:04:05.0000000"))

	case strings.HasPrefix(typename, "timestamp"):
		return fmt.Sprintf("%s", v.Format("2006/01/02 15:04"))

	case strings.HasPrefix(typename, "date"):
		return fmt.Sprintf("%s", v.Format("2006/01/02 15:04:05"))

	default:
		return fmt.Sprintf("%s", v.Format("2006/01/02 15:04:05.000"))
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
		return fmt.Sprintf("%s", (*v).Format("2006/01/02 15:04:05.000"))
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

	default:
		return "NULL"
	}
	return "NULL"
}
