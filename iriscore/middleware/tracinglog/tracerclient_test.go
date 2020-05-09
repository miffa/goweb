package tracinglog

import (
	"context"
	"fmt"
	"testing"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"goweb/iriscore/libs/gorequest"
	"goweb/iriscore/libs/mitracer"
)

//curl -XPOST http://10.104.108.228:8080/demopost

func TestTracingClient(t *testing.T) {

	ihttpctx := context.Background()
	//opentracing.ContextWithSpan(ctx, span)

	// make new span
	serspan, serctx := opentracing.StartSpanFromContextWithTracer(ihttpctx,
		mitracer.New(),
		"heraGotest",
	)
	//serspan.SetTag("", "")
	//serspan.SetBaggageItem("", "")
	serspan.SetTag("http.method", "GET")
	serspan.SetTag("http.url", "-----")
	serspan.SetTag("http.host", "nicai")
	serspan.SetTag("http.peer", "127.0.0.1")
	client := gorequest.New().Get("http://10.104.108.228:8088/demoget").
		WithContext(serctx).
		Timeout(5 * time.Second) //.

	tm.tracer = mitracer.New()
	PutMiSpanToHeader(serspan, client.Header)

	_, body, ierrors := client.End()

	if len(ierrors) != 0 {
		for _, err := range ierrors {
			t.Error(err)
		}
	}
	serspan.Finish()
	fmt.Println(body)
	fmt.Println(serspan)
}
