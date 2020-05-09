package esquery

import (
	"errors"
	"strings"
	"time"

	"github.com/olivere/elastic"
)

const (
	ES_DEPARTMENT  = "kubernetes.labels.tpaas-ns"
	ES_PROJECT     = "kubernetes.labels.tpaas-app"
	ES_SERVICE     = "kubernetes.labels.tpaas-controller-kind"
	ES_POD         = "kubernetes.pod_name"
	ES_CONTAINER   = "kubernetes.container_name"
	ES_SERVICE_APP = "kubernetes.labels.tpaas-name"
	ES_MESSAGE     = "message"

	DOC_TIMESTAMP = "@timestamp"
	DOC_SERIALNO  = "_doc"
	//DOC_TYPE      = "_doc"
	DOC_TYPE = "flb_type"

	T_SCROLL_UP   = false
	T_SCROLL_DOWN = true
)

var (
	ES_CFG_ADDR_ERR         = errors.New("es does not set address")
	ES_QUERY_CONTEXTPOS_ERR = errors.New("es searchafter pos not found")
	Debug                   = true
	EL_LOG_CTX_ALL          = []bool{T_SCROLL_UP, T_SCROLL_DOWN}
	EL_LOG_CTX_UP           = []bool{T_SCROLL_UP}
	EL_LOG_CTX_DOWN         = []bool{T_SCROLL_DOWN}
)

////////////////////////
//// logquerying connection
////////////////////////

//  client connection pool
var (
	EsPool = EsCliConnPool{
		pool: make(map[string]*EsCliConn),
	}
)

func ESPool() *EsCliConnPool {
	return &EsPool
}

type EsCliConnPool struct {
	pool map[string]*EsCliConn
}

func (ep *EsCliConnPool) AddConn(ec *EsCliConn) {
	keys := strings.Join(ec.addrs, "_")
	ep.pool[keys] = ec
}

func (ep *EsCliConnPool) RemoveConn(ec *EsCliConn) {
	keys := strings.Join(ec.addrs, "_")
	delete(ep.pool, keys)
}

// ClientOptionFunc
type ClientOptionFunc func(*EsCliConn) error

func SetURL(ad []string) ClientOptionFunc {
	return func(c *EsCliConn) error {
		c.addrs = ad
		return nil
	}
}

func SetBaseAuth(u, p string) ClientOptionFunc {
	return func(c *EsCliConn) error {
		c.user = u
		c.passwd = p
		return nil
	}
}

func SetTimeout(sec time.Duration) ClientOptionFunc {
	return func(c *EsCliConn) error {
		c.tmout = sec
		return nil
	}
}

// es client conn
type EsCliConn struct {
	addrs  []string
	passwd string
	user   string
	tmout  time.Duration
	cli    *elastic.Client
}

func NewEsClient(opts ...ClientOptionFunc) (*EsCliConn, error) {
	esconn := &EsCliConn{}

	for _, opt := range opts {
		if err := opt(esconn); err != nil {
			return nil, err
		}
	}

	if len(esconn.addrs) == 0 {
		return nil, ES_CFG_ADDR_ERR
	}

	if esconn.tmout < 1*time.Second {
		esconn.tmout = 30 * time.Second
	}

	escli, err := elastic.NewClient(elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetGzip(true),
		elastic.SetURL(esconn.addrs...),
	)
	if err != nil {
		return nil, err
	}
	esconn.cli = escli
	return esconn, nil
}

func (ec *EsCliConn) Close() error {
	ec.cli.Stop()
	return nil
}
