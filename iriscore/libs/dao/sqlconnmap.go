package dao

import (
	"database/sql"
	"fmt"
	"time"

	"goweb/iriscore/libs/dao/dbinfo"
	log "goweb/iriscore/libs/logrus"

	"github.com/go-gorp/gorp"
)

type MakeDbConn func() DbConn

var (
	DbConnFactory map[string]MakeDbConn = nil
)

func RegisterFactory(k string, v MakeDbConn) {
	if DbConnFactory == nil {
		DbConnFactory = make(map[string]MakeDbConn)
	}
	DbConnFactory[k] = v
}

type DbConn interface {
	Makedatasource(ip, db, user, pwd string) string
	Init(dataaddr string) error
	SelectTableForDb(dbname string) ([]dbinfo.TDBTab, error)
	SelectColumnForTable(dbname, tabname string) ([]dbinfo.TDBCol, error)
	Execute(sqlstmt string) (int, error)
	SelectWithOrm(result interface{}, sqlstmt string) error
	SelectValues(sqlstmt string) ([][]string, []string, error)
	Select(sqlstmt string) ([][]string, []string, []int, error)
	Killsql(db, sql string) (bool, error)
	BatchExecute(sqls []string) error
	GetDbConn() *sql.DB
	GetDbObj() *gorp.DbMap
	Ping() error
	Close()
}

type SqlConn struct {
	dbconn     DbConn
	ds         string
	keepalived bool
	dbname     string
}

func NewDbConn(ipaddr, db, user, rose string, typestr string) (*SqlConn, error) {
	sqlconn := &SqlConn{}
	err := sqlconn.InitConn(ipaddr, db, user, rose, typestr)
	if err != nil {
		return nil, err
	}
	return sqlconn, nil
}

func (c *SqlConn) GetDbObj() *gorp.DbMap {
	return c.dbconn.GetDbObj()
}

func (c *SqlConn) InitConn(ipaddr, db, user, rose string, typestr string) error {

	/*
		switch typestr {
		case "mysql", "MySQL", "MYSQL":
			c.dbconn = &mysqldao.MySqlConn{}
			c.ds = c.dbconn.Makedatasource(ipaddr, db, user, rose)

		case "tidb", "TIDB", "TiDB":
			c.dbconn = &mysqldao.MySqlConn{}
			c.ds = c.dbconn.Makedatasource(ipaddr, db, user, rose)

		case "sqlserver", "mssql", "SQL Server":
			c.dbconn = &mssqldao.MsSqlConn{}
			c.ds = c.dbconn.Makedatasource(ipaddr, db, user, rose)

		case "oracle", "Oracle":
			c.dbconn = &oracledao.OracleConn{}
			c.ds = c.dbconn.Makedatasource(ipaddr, db, user, rose)
		default:
			return fmt.Errorf("%s not support", typestr)
		}*/

	birth, ok := DbConnFactory[typestr]
	if !ok {
		return fmt.Errorf("%s not support", typestr)
	}
	c.dbconn = birth()
	c.ds = c.dbconn.Makedatasource(ipaddr, db, user, rose)

	err := c.connect()
	if err != nil {
		return err
	}
	c.dbname = db
	return nil
}

func (c *SqlConn) InitConnWithKeepAlived(ipaddr, db, user, rose string, typestr string) error {
	err := c.InitConn(ipaddr, db, user, rose, typestr)
	if err != nil {
		return nil
	}
	go c.AmiAlive()
	return nil
}

func (c *SqlConn) connect() error {
	return c.dbconn.Init(c.ds)
}

func (c *SqlConn) Close() {
	if c.keepalived {
		c.keepalived = false
	} else {
		c.dbconn.Close()
	}
}

func (c *SqlConn) AmiAlive() error {
	c.keepalived = true
	ticker := time.NewTicker(10 * time.Second)

	var err error
	for c.keepalived {
		select {
		case <-ticker.C:
			if err = c.dbconn.Ping(); err != nil {
				//c.connect()
			}
		}
	}
	c.dbconn.Close()
	return nil
}

func (c *SqlConn) Killsql(db, sql string) (bool, error) {
	return c.dbconn.Killsql(db, sql)
}

func (c *SqlConn) ExecGrpStmt(stmt []string) error {
	err := c.dbconn.BatchExecute(stmt)
	if err != nil {
		return err
	}
	return nil
}
func (c *SqlConn) ExecStmt(stmt string) (int, error) {
	ret, err := c.dbconn.Execute(stmt)
	if err != nil {
		return -1, err
	}
	if ret == 0 {
		log.Warnf("exec [%s] Affected rows is 0", stmt)
		return -1, nil
	}
	return ret, nil
}

func (c *SqlConn) QueryValuesWithType(stmt string) ([][]string, []string, []int, error) {
	ret, cols, tys, err := c.dbconn.Select(stmt)
	if err != nil {
		return nil, nil, nil, err
	}
	return ret, cols, tys, nil
}

func (c *SqlConn) QueryValues(stmt string) ([][]string, []string, error) {
	ret, cols, err := c.dbconn.SelectValues(stmt)
	if err != nil {
		return nil, nil, err
	}
	return ret, cols, nil
}

func (c *SqlConn) SelectWithOrm(result interface{}, stmt string) error {
	err := c.dbconn.SelectWithOrm(result, stmt)
	if err != nil {
		return err
	}
	return nil
}

func (c *SqlConn) Columns(table string) ([]dbinfo.TDBCol, error) {
	return c.dbconn.SelectColumnForTable(c.dbname, table)
}

func (c *SqlConn) Tables() ([]dbinfo.TDBTab, error) {
	return c.dbconn.SelectTableForDb(c.dbname)
}
