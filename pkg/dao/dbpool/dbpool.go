package dbpool

import (
	"iris/pkg/config"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var dbpool *TpaasDbPool

type TpaasDbPool struct {
	DB *gorm.DB
}

func Pool() *TpaasDbPool {
	if dbpool == nil {
		dbpool = new(TpaasDbPool)
	}
	return dbpool
}

func (d *TpaasDbPool) Init(cfg *config.TpaasConfig) error {
	var err error
	d.DB, err = gorm.Open("mysql", cfg.GetString("mysql.datasource"))
	if err != nil {
		return err
	}
	maxopen := cfg.GetInt("mysql.max_open_conn")
	if maxopen == 0 {
		maxopen = 50
	}
	d.DB.DB().SetMaxOpenConns(maxopen)

	maxidle := cfg.GetInt("mysql.max_idle_conn")
	if maxidle == 0 {
		maxidle = 5
	}
	d.DB.DB().SetMaxIdleConns(maxidle)

	maxlifetime := cfg.GetInt("mysql.max_conn_lifetime")
	if maxlifetime != 0 {
		d.DB.DB().SetConnMaxLifetime(time.Duration(maxlifetime) * time.Minute)
	}

	maxidletime := cfg.GetInt("mysql.max_conn_idletime")
	if maxidletime != 0 {
		d.DB.DB().SetConnMaxIdleTime(time.Duration(maxidletime) * time.Minute)
	}
	return nil
}

func (d *TpaasDbPool) Close() error {
	return d.DB.Close()
}

func NewDbPoolWithDsn(dsn string) (*TpaasDbPool, error) {
	dbpool := new(TpaasDbPool)
	var err error
	dbpool.DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return dbpool, nil

}
