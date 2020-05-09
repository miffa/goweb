package dbpool

import (
	"goweb/iriscore/config"
	"goweb/iriscore/iocgo"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var dbpool *TpaasDbPool

type TpaasDbPool struct {
	db *gorm.DB
}

func Pool() *TpaasDbPool {
	if dbpool == nil {
		dbpool = new(TpaasDbPool)
	}
	return dbpool
}

func (d *TpaasDbPool) Init(cfg *config.TpaasConfig) error {
	var err error
	d.db, err = gorm.Open("mysql", cfg.GetString("mysql.datasource"))
	if err != nil {
		return err
	}
	return nil
}

func (d *TpaasDbPool) Close() error {
	return d.db.Close()
}

func NewDbPoolWithDsn(dsn string) (*TpaasDbPool, error) {
	dbpool := new(TpaasDbPool)
	var err error
	dbpool.db, err = gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return dbpool, nil

}

func init() {
	iocgo.Register("db datasource pool", Pool())
}
