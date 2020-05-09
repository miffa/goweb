package tracinglog

import (
	"fmt"
	"net/http"
	"strings"
	"goweb/iriscore/config"
	"goweb/iriscore/define"
	"goweb/iriscore/iocgo"
	"goweb/iriscore/version"

	"goweb/iriscore/libs/mitracer"

	iris "github.com/kataras/iris/v12"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

var tm *TracerMgr = &TracerMgr{}

func init() {
	iocgo.Register("tracerlog", tm)
}

type TracerMgr struct {
	tracer       opentracing.Tracer
	tracerLogger *DeckLogger
}

func (t *TracerMgr) Init(cfg *config.TpaasConfig) error {
	t.tracer = mitracer.New()
	if !cfg.IsSet("common.tracer_log") {
		return errors.New("no common.tracer_log in config file")
	}
	t.tracerLogger = NewDeckLogger(cfg.GetString("common.tracer_log"))
	return nil
}

func (t *TracerMgr) Close() error {
	return t.tracerLogger.Close()
}

func Tracing(ctx iris.Context) {
	carrier := opentracing.HTTPHeadersCarrier(ctx.Request().Header)
	clientContext, _ := tm.tracer.Extract(opentracing.HTTPHeaders, carrier)

	var opts []opentracing.StartSpanOption
	operationName := strings.Replace(ctx.Path(), "/", "_", -1)
	opts = append(opts, opentracing.ChildOf(clientContext))
	serspan := tm.tracer.StartSpan(operationName, opts...)
	serctx := opentracing.ContextWithSpan(ctx.Request().Context(), serspan)

	// make new span
	//serspan, serctx := opentracing.StartSpanFromContextWithTracer(ihttpctx,
	//	tm.tracer,
	//	strings.Replace(ctx.Path(), "/", "_", -1),
	//)
	//serspan.SetTag("", "")
	//serspan.SetBaggageItem("", "")
	serspan.SetTag("http.method", ctx.Method())
	serspan.SetTag("http.url", ctx.Path())
	serspan.SetTag("http.host", ctx.Host())
	serspan.SetTag("http.peer", ctx.RemoteAddr())
	serspan.SetTag("http.remote_addr", ctx.Request().RemoteAddr)
	serspan.SetTag("service.name", version.Service)
	serspan.SetTag("service.versinon", version.Version)

	//  set span ctx with request to iris ctx,
	//then we can send ctx to another system
	ctx.Request().WithContext(serctx)

	// set in ctx.Values
	ctx.Values().Set(define.CTX_VALUE_SPAN_KEY, serspan)
	ctx.Next()
}

func FinishSpan(ctx iris.Context) {
	var status interface{}
	st := ctx.Values().Get(define.CTX_RESP_STS_KEY)
	if st != nil {
		status = st
	} else {
		status = "-"
	}

	// get span from ctx
	ptrspan := ctx.Values().Get(define.CTX_VALUE_SPAN_KEY)
	if ptrspan == nil {
		return
	}
	//span, ok := ptrspan.(opentracing.Span)
	span, ok := ptrspan.(*mitracer.MiSpan)
	if !ok {
		return
	}

	span.LogKV("response", status)

	// span finish
	span.Finish()

	// log to logger
	//tm.tracerLogger.WithField("service", version.Service).WithField("version", version.Version).Info(span)
	tm.tracerLogger.WithField("service", version.Service).
		WithField("version", version.Version).
		WithField("traceId", span.SpanContext.TraceID).
		WithField("spanId", span.SpanContext.SpanID).
		WithField("parentId", span.ParentID).
		WithField("startTime", span.StartTime.UnixNano()/1000000).
		WithField("finishTime", span.FinishTime.UnixNano()/1000000).
		WithField("costTime", (span.FinishTime.UnixNano()-span.StartTime.UnixNano())/1000000).
		WithField("sampled", span.SpanContext.Sampled).
		WithField("tags", span.Tags()).
		//WithField("log", span.Logs()).
		WithField("name", span.OperationName).Info(status)
	//WithField("", span.   ).
	//WithField("", span.   ).

	tm.tracer.(*mitracer.MiTracer).Reset()
}

func PutSpanToHeader(ctx iris.Context, h http.Header) error {
	serspan := ctx.Values().Get(define.CTX_VALUE_SPAN_KEY)
	if serspan == nil {
		return fmt.Errorf("not found span")
	}
	span, ok := serspan.(opentracing.Span)
	if !ok {
		return fmt.Errorf("err type span")
	}

	err := tm.tracer.Inject(span.Context(), opentracing.HTTPHeaders, h)
	if err != nil {
		return err
	}

	return nil
}

func PutMiSpanToHeader(span opentracing.Span, h http.Header) error {
	err := tm.tracer.Inject(span.Context(), opentracing.HTTPHeaders, h)
	if err != nil {
		return err
	}

	return nil
}
