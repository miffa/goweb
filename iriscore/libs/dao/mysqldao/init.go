package mysqldao

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

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	//_ "github.com/ziutek/mymysql"
	"goweb/iriscore/libs/dao"
	dbtype "goweb/iriscore/libs/dao/dbnulltype"
	"goweb/iriscore/libs/dao/dbtimeout"
)

const (
	DB_TYPE           = "mysql"
	TABLE_SELECT_SPEC = `
		SELECT TABLE_NAME,TABLE_ROWS,DATA_LENGTH/1024/1024 "used_mb"
		FROM
		INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = ? ORDER BY TABLE_ROWS DESC; `

	COLUMN_SELECT_SPEC = `
		SELECT  c.COLUMN_NAME,c.COLUMN_TYPE,
		CASE c.extra WHEN 'auto_increment' THEN 'YES' ELSE 'NO' END AS 'is_identity',
		CASE kcu.CONSTRAINT_NAME WHEN 'PRIMARY' THEN 'YES' ELSE 'NO' END AS 'IS_PRIMARY',
		c.IS_NULLABLE,c.COLUMN_COMMENT,
		case  when COLUMN_DEFAULT  is null then 'Null' else COLUMN_DEFAULT end as 'column_default' 
		FROM
		INFORMATION_SCHEMA.COLUMNS c
		 LEFT JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu 
		 ON kcu.TABLE_SCHEMA = c.TABLE_SCHEMA AND kcu.TABLE_NAME = c.TABLE_NAME AND kcu.COLUMN_NAME = c.COLUMN_NAME
		 WHERE c.table_schema = ? AND c.table_name = ?; `
)

var (
	ERRNilDB    = errors.New("nil mysql conn")
	ERRNilObj   = errors.New("nil obj")
	ERRDBConn   = errors.New("conn db err")
	ERREXISTOBJ = errors.New("obj has already exits")

	ZeroId = 0
)

var (
	DBMSMgr *MySqlConn
)

func GenMysqlDbconn() dao.DbConn {
	return &MySqlConn{}
}

func init() {
	dao.RegisterFactory("MySQL", GenMysqlDbconn)
	dao.RegisterFactory("mysql", GenMysqlDbconn)
	dao.RegisterFactory("MYSQL", GenMysqlDbconn)
	dao.RegisterFactory("TIDB", GenMysqlDbconn)
	dao.RegisterFactory("TiDB", GenMysqlDbconn)
	dao.RegisterFactory("tidb", GenMysqlDbconn)
}

func InitDbConn(dbaddr string) error {
	DBMSMgr = &MySqlConn{}
	return DBMSMgr.Init(dbaddr)
}

func CloseDbConn() {
	DBMSMgr.Close()
}

type MySqlConn struct {
	DBObj    *gorp.DbMap
	DBConn   *sql.DB
	dataaddr string //"root:root@tcp(10.3.100.64:3306)/groupdb?parseTime=true&loc=Local&charset=utf8"
}

func (c *MySqlConn) Init(dataaddr string) error {
	var err error
	c.dataaddr = dataaddr
	c.DBConn, err = sql.Open(DB_TYPE, dataaddr)
	if err != nil {
		return err
	}
	c.DBObj = &gorp.DbMap{Db: c.DBConn, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}

	return nil
}

func (c *MySqlConn) Execute(sqlstmt string) (int, error) {
	if rets, err := c.DBObj.Exec(sqlstmt); err != nil {
		return 0, err
	} else {
		rows, _ := rets.RowsAffected()
		return int(rows), nil
	}
}

func (c *MySqlConn) BatchExecute(sqls []string) error {

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

func (c *MySqlConn) SelectWithOrm(result interface{}, sqlstmt string) error {
	_, err := c.DBObj.Select(result, sqlstmt)
	return err
}

type columnType struct {
	Ty    int
	valid bool
}

func (c *MySqlConn) Select(sqlstmt string) ([][]string, []string, []int, error) {
	var ret [][]string
	//rows, err := c.DBConn.Query(sqlstmt)
	ctx, cancel := context.WithTimeout(context.Background(), dbtimeout.CONNECTION_TIMEOUT)
	defer cancel()
	rows, err := c.DBConn.QueryContext(ctx, sqlstmt)
	if err != nil {
		return nil, nil, nil, err
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, nil, err
	}

	if len(cols) == 0 {
		return nil, nil, nil, fmt.Errorf("nil columns")
	}
	coltys, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, nil, err
	}
	var colinfo []*columnType
	for _, cv := range coltys {
		log.Debugf("xxxxxxx column:%s  golang_type:%s db_typ:%s", cv.Name(), cv.ScanType().Name(), cv.DatabaseTypeName())
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
		for i, _ := range vals {
			datastr, ty := TostringType(vals[i].(*interface{}), strings.ToLower(coltys[i].DatabaseTypeName()))
			strv = append(strv, datastr)
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

func (c *MySqlConn) SelectValues(sqlstmt string) ([][]string, []string, error) {
	var ret [][]string
	//rows, err := c.DBConn.Query(sqlstmt)
	ctx, cancel := context.WithTimeout(context.Background(), dbtimeout.CONNECTION_TIMEOUT)
	defer cancel()
	rows, err := c.DBConn.QueryContext(ctx, sqlstmt)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	if len(cols) == 0 {
		return nil, nil, fmt.Errorf("nil columns")
	}
	coltys, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}
	for _, cv := range coltys {
		log.Debugf("xxxxxxx column:%s  golang_type:%s db_typ:%s", cv.Name(), cv.ScanType().Name(), cv.DatabaseTypeName())
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
		for i, _ := range vals {
			strv = append(strv, Tostring(vals[i].(*interface{}), strings.ToLower(coltys[i].DatabaseTypeName())))
		}
		ret = append(ret, strv)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}
	return ret, cols, nil
}

type Processlist struct {
	Id   int    `db:"id"`
	Host string `db:"host"`
	Db   string `db:"db"`
	Info string `db:"info"`
	Time string `db:"time"`
}

func (c *MySqlConn) Killsql(db, sql string) (bool, error) {
	var ret []Processlist
	err := c.SelectWithOrm(&ret, `
	SELECT id,host,db,time,info, time
	FROM information_schema.PROCESSLIST
	WHERE user = 'admintools' AND time >= 99;
	`)
	if err != nil {
		return false, err
	}
	if len(ret) == 0 {
		log.Warnf("kill sql::PROCESSLIST is empty")
		return false, nil
	}

	pos := 0
	found := false
	for pp, pi := range ret {
		log.Debugf("kill sql::find sql job:[%s]: %s", sql, pi.Info)
		if db == "" {
			if pi.Info == sql {
				pos = pp
				found = true
				log.Infof("kill sql::found sql job:[%s]:%d", sql, pi.Id)
				break
			}
		} else {
			if pi.Info == sql && pi.Db == db {
				pos = pp
				found = true
				log.Infof("kill sql::found sql job:[%s]:%d", sql, pi.Id)
				break
			}
		}
	}

	if !found {
		log.Warnf("kill sql::not found sql job:[%s]", sql)
		return false, nil
	}

	_, err = c.Execute(fmt.Sprintf("kill %d", ret[pos].Id))
	if err != nil {
		log.Warnf("kill sql::not found sql job:[%s]", sql)
		return false, err
	}
	//if iret == 0 {
	//	log.Warnf("exec [kill %s] Affected rows is 0", sql)
	//	return false, nil
	//}

	log.Infof("kill sql::found sql job:[%s] and kill ok", sql)
	return true, nil
}

//
func (c *MySqlConn) Ping() error {
	return c.DBConn.Ping()
}

func (c *MySqlConn) GetDbConn() *sql.DB {
	return c.DBConn
}

//
func (c *MySqlConn) GetDbObj() *gorp.DbMap {
	return c.DBObj
}

//
func (c *MySqlConn) Close() {
	c.DBConn.Close()
}

const (
	DATA_SOURCE_DEMO = "dbms_user:dbms_pdw@tcp(10.101.xx.xx:3306)/db_cmdb?parseTime=true&loc=Local&charset=utf8"
	DATA_SOURCE_L    = "("
	DATA_SOURCE_R    = ")"
	DATA_SOURCE_P    = "/"
	DATA_SOURCE_SUF  = "?parseTime=true&loc=Local&charset=utf8&timeout=30s&readTimeout=3600s"
	DB_SYS_DB        = "INFORMATION_SCHEMA"
	CONN_TIMEOUT     = "timeout=30s"
	READ_TIMEOUT     = "readTimeout=3000s"
)

func (c *MySqlConn) Makedatasource(ipport, dbname string, user, pass string) string {
	return fmt.Sprintf("%s:%s@tcp%s%s%s%s%s%s", user, pass, DATA_SOURCE_L, ipport, DATA_SOURCE_R, DATA_SOURCE_P, dbname, DATA_SOURCE_SUF)
}

func (c *MySqlConn) Tostring(av interface{}, typeinfo string) string {
	return ""
}

func (c *MySqlConn) SelectTableForDb(dbname string) ([]proto.TDBTab, error) {
	begin := time.Now()
	var tabs []proto.TDBTab
	if _, err := c.DBObj.Select(&tabs, TABLE_SELECT_SPEC, dbname); err != nil {
		return nil, err
	}

	log.Debugf("SelectTableForDb ok costtime:%s", time.Now().Sub(begin).String())
	return tabs, nil
}

func (c *MySqlConn) SelectColumnForTable(dbname, tabname string) ([]proto.TDBCol, error) {

	begin := time.Now()
	var cols []proto.TDBCol
	if _, err := c.DBObj.Select(&cols, COLUMN_SELECT_SPEC, dbname, tabname); err != nil {
		return nil, err
	}

	log.Debugf("SelectColumnForTable ok costtime:%s", time.Now().Sub(begin).String())
	return cols, nil
}

//////////////////////////////////////
func binaryDataConv(data []byte, typename string) string {
	typename = strings.ToLower(typename)
	switch {
	case strings.HasPrefix(typename, "varchar"):
		fallthrough
	case strings.HasSuffix(typename, "text"):
		return string(data)
	case strings.HasSuffix(typename, "blob"):
		return string(data)
		//return "(BLOB)"
	case strings.HasPrefix(typename, "bit"):
		return "0x" + hex.EncodeToString(data)
	case strings.HasPrefix(typename, "geometry"):
		return "0x" + hex.EncodeToString(data)
	default:
		//return "0x" + hex.EncodeToString(data) + "(" + string(data) + ")"
		return string(data)
	}

	return ""
}

func timeDataConv(v time.Time, typename string) string {
	typename = strings.ToLower(typename)
	switch {
	case strings.HasPrefix(typename, "timestamp"):
		return fmt.Sprintf("%s", v.Format("20060102150405"))

	case strings.HasPrefix(typename, "datetime"):
		return fmt.Sprintf("%s", v.Format("2006-01-02 15:04:05"))

	case strings.HasPrefix(typename, "date"):
		return fmt.Sprintf("%s", v.Format("2006-01-02"))

	case strings.HasPrefix(typename, "year"):
		return fmt.Sprintf("%s", v.Format("2006"))

	case strings.HasPrefix(typename, "time"):
		return fmt.Sprintf("%s", v.Format("15:04:05"))
	default:
		return fmt.Sprintf("%s", v.Format("2006-01-02 15:04:05"))
	}

	return ""
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

	default:
		return "NULL"
	}
	return "NULL"
}
