package esquery

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
	"tpaas/src/logforce/proto"

	"github.com/elastic/go-elasticsearch"
	"github.com/sirupsen/logrus"

	//"github.com/olivere/elastic"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}
func TestDemo(t *testing.T) {

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://172.27.139.80:29200/",
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logrus.Debugf("error new esclihahaent %v", err)
		return
	}

	res, err := es.Info()
	if err != nil {
		logrus.Debugf("error new esclihahaent %v", err)
		return
	}
	logrus.Debugf("es.info:%v\n", res)
	logrus.Debugf("this is a test\n")

	logrus.Debug("-------")
	logrus.Debug("elastic lib")
	// es client
	esclihaha, err := elastic.NewClient(elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetGzip(true),
		elastic.SetURL("http://172.27.139.80:29200"))
	if err != nil {
		// Handle error
		logrus.Debug(err)
		return
	}
	defer esclihaha.Stop()
	logrus.Debug("-------ping ")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pingres, num, err := esclihaha.Ping("http://172.27.139.80:29200").Do(ctx)
	if err != nil {
		logrus.Debugf("ping err %v\n", err)
		return
	}
	logrus.Debugf("ping service %d:%v\n", num, pingres)
	time.Sleep(5 * time.Second)

	logrus.Debug("-------index exists")
	// index exist
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, err := esclihaha.IndexExists("logstash-2019.11.13").Do(ctx)
	if err != nil {
		// Handle error
		logrus.Debug("xxxx", err)
		return
	}
	if !exists {
		// Index does not exist yet.
		logrus.Debug("not exist index")
		return
	}
	logrus.Debug("-------mapping")

	// get mapping
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mapres, err := elastic.NewGetFieldMappingService(esclihaha).
		Index("logstash-2019.11.15").
		Pretty(true).Do(ctx)
	if err != nil {
		logrus.Debug("mapping", err)
		return
	}
	data, _ := json.Marshal(mapres)
	logrus.Debug(string(data))

	// get search data
	//// search service
	logrus.Debug("-------search")

	// project query
	//pjquery := elastic.NewBoolQuery()

	// query string
	bq := elastic.NewBoolQuery()

	tq := elastic.NewTermQuery("kubernetes.labels.app", "nginx")
	mq := elastic.NewMatchQuery("kubernetes.pod_name", "nginx-b4588bf76-h6xnq")
	kmq := elastic.NewMatchQuery("message", "ApacheBench")
	rq := elastic.NewRangeQuery("@timestamp")
	// ms
	rq.Gte(10000).Lt(1583635096000)
	// string time
	//rq.Gte("2019-11-11").Lte("2019-11-13").TimeZone("Asia/Shanghai")
	bq.Must(tq, mq, kmq, rq)

	// histogram
	//sumt := elastic.NewHistogramAggregation().Interval(30000 /*float64*/).Field("@timestamp")

	//date histogram
	// yyyyMMdd'T'HHmmss
	datesumt := elastic.NewDateHistogramAggregation().
		Field("@timestamp").
		Interval("1m").
		Format("yyyy-MM-dd HH:mm:ss").
		TimeZone("Asia/Shanghai").
		MinDocCount(0)

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	searchsv := elastic.NewSearchService(esclihaha).
		Index("logstash-*").
		From(0).
		Size(10).
		//Query(bq).Aggregation("@timestamp", sumt)
		Query(bq).Aggregation("@timestamp", datesumt)

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	retdata, err := searchsv.Do(ctx)
	if err != nil {
		logrus.Debug("search error:", err)
		return
	}
	logrus.Debug("esdata:", retdata.Hits.TotalHits)
	for p, dd := range retdata.Hits.Hits {
		logrus.Debugf("hit:%d:%s\n", p, ToJson(dd.Source))
	}
	//aggs, b := retdata.Aggregations.Histogram("@timestamp")
	aggs, b := retdata.Aggregations.DateHistogram("@timestamp")
	if b {
		logrus.Debugf("esdata histogram: %T", aggs.Buckets)

		for _, data := range aggs.Buckets {
			dc := data.DocCount
			k := data.Key
			ks := data.KeyAsString
			logrus.Debugf("esdata histogram:  %#v   %#v:%#v  %#v\n", dc, int64(k), k, *ks)
		}
	} else {
		logrus.Debugf("histogram not found\n")
	}

	logrus.Debug("-------")
}

func TestSearchPos(t *testing.T) {
	logrus.Debugf("==========================================================")
	ec, err := NewEsClient(SetURL([]string{"http://172.27.139.80:29200/"}),
		SetTimeout(30*time.Second))
	if err != nil {
		t.Errorf("create es clooent err:%v", err)
	}
	defer ec.Close()

	/*args := &proto.EsSearchAfterPosArg{
		Departments: "77",
		Projects:    "78",
		Kinds:       "deployment",
		Pods:        "nginx-b4588bf76-h6xnq",
		Containers:  "container1",
		KeyWord:     []string{"ApacheBench", "GET"},
		Indexes:     []string{"logstash-*"},
		StartTime:   100000,
		EndTime:     1583635096000,
		ID:          "_oTxY24BKr004d9-GPsj",
	}*/
	args := &proto.EsSearchAfterArg{}
	args.Departments = 77
	args.Projects = 78
	args.Kinds = "deployment"
	args.Pods = "nginx-b4588bf76-h6xnq"
	args.Containers = "container1"
	args.Indexes = []string{"logstash-*"}
	args.StartTime = 1000000
	args.EndTime = 1583635096000
	args.ID = "_4TxY24BKr004d9-GPsj"
	args.Size = 5
	//pos, err := ec.queryLogContextPos(args)
	//if err != nil {
	//	t.Errorf("QueryLogContextPos err:%v", err)
	//}

	//logrus.Infof("searafter pos:%s  %#v", ToJson(pos), pos)

	//ir, err := ec.searchAfterContext(args, pos)
	ir, err := ec.QueryLogContext(args)
	if err != nil {
		t.Errorf("=====:%v", err)
	}

	logrus.Debugf("=======:%#v\n", ir)
	for _, v := range ir {
		if v.Order {
			logrus.Debugf("asc   ascascascascascascasc==%d", len(v.Logs))
		} else {
			logrus.Debugf("desc  descdescdescdescdescdescdesc==%d", len(v.Logs))

		}
		for _, vv := range v.Logs {
			logrus.Debugf("log_detail:%#v:%#v", vv.Sort, string(*(vv.Source)))
		}
	}
}

func TestSearchContextPage(t *testing.T) {
	logrus.Debugf("==========================================================")
	ec, err := NewEsClient(SetURL([]string{"http://172.27.139.80:29200/"}),
		SetTimeout(30*time.Second))
	if err != nil {
		t.Errorf("create es clooent err:%v", err)
	}
	defer ec.Close()

	args := proto.EsSearchAfterArg{}
	allargs := &proto.EsSearchAfterPageArg{}
	args.Departments = 77
	args.Projects = 78
	args.Kinds = "deployment"
	args.Pods = "nginx-b4588bf76-h6xnq"
	args.Containers = "container1"
	args.Indexes = []string{"logstash-*"}
	args.StartTime = 1000000
	args.EndTime = 1583635096000
	args.Size = 5
	allargs.Sort = []interface{}{1573634770355, 200771}
	allargs.SortAsc = true
	allargs.EsSearchAfterArg = args

	ir, err := ec.QueryLogContextNextPage(allargs)
	if err != nil {
		t.Errorf("=====:%v", err)

	}

	logrus.Debugf("=======:%#v\n", ir)
	for _, v := range ir {
		if v.Order {

			logrus.Debugf("asc==%v", v.Logs)
		} else {
			logrus.Debugf("desc==%v", v.Logs)

		}
		for _, vv := range v.Logs {
			logrus.Debugf("log_detail:%#v:%#v", vv.Sort, string(*(vv.Source)))
		}
	}
}

func ExampleHello() {
	fmt.Println("Hello")
	// Output: Hello
}
