package oracledao

import (
	"fmt"
	"testing"
	//oracle driver
)

var debug = fmt.Println
var debugf = fmt.Printf

func TestInit(t *testing.T) {
	addr := "10.104.20.121:1521"
	//addr := "10.104.20.121"
	//addr := ""
	user := "dbms_q"
	pwd := "okkkkk"
	orc := &OracleConn{}
	dsn := orc.Makedatasource(addr, "HEC3UAT::HEC3UAT", user, pwd)
	debug(dsn)
	//dsn = `(DESCRIPTION = (ADDRESS = (PROTOCOL = TCP)(HOST = 10.104.20.121)(PORT = 1521)) (CONNECT_DATA = (SERVER = DEDICATED) (SERVICE_NAME = hec3uat)))`
	debug(dsn)
	err := orc.Init(dsn)
	if err != nil {
		t.Errorf("error:%v", err)
		return
	}
	debugf("init ok\n")

	err = orc.Ping()
	if err != nil {
		t.Errorf("ping error:%v", err)
		return
	}
	debug("ping")

	data, err := orc.SelectTableForDb("HEC3UAT")
	if err != nil {
		t.Errorf("error:%v", err)
		return
	}
	for p, o := range data {
		debugf("    %d:  %s -- %f rows %f MB  \n", p, o.TableName, o.RowCount.Float64, o.TableMb.Float64)
	}

	//UCAR_NEWCAR_BUSINESS_DETAILS
	cols, err := orc.SelectColumnForTable("HEC3UAT", "UCAR_NEWCAR_BUSINESS_DETAILS")
	if err != nil {
		t.Errorf("error:%v", err)
		return
	}

	for o, p := range cols {
		debugf("    %d:    %#v \n", o, p)
	}

	// ok

}
