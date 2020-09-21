package define

import "fmt"

const (
	InnerServerErr = "服务器内部错误"
)

type TpaasError interface {
	// error info to response.Detail
	Error() string
	// return code to response.Code
	RetCode() int
	// response message to response.Msg
	Message() string
}

///////////

var _ TpaasError = new(errNotFound)
var _ TpaasError = new(errSvcNotFound)
var _ TpaasError = new(errToken)
var _ TpaasError = new(errInnerServer)
var _ TpaasError = new(errSvcTmout)

var _ error = new(errNotFound)
var _ error = new(errSvcNotFound)
var _ error = new(errToken)
var _ error = new(errInnerServer)
var _ error = new(errSvcTmout)

type errSvcTmout struct {
	Key  string
	Code int
}

func ErrSvcTmout(servicename string) error {
	return &errSvcTmout{Key: servicename, Code: ST_SERVICE_TIMEOUT}
}

func (e *errSvcTmout) Error() string {
	return fmt.Sprintf("service[%s] is timeout", e.Key)
}

func (e *errSvcTmout) RetCode() int {
	return ST_SERVICE_TIMEOUT
}

func (e *errSvcTmout) Message() string {
	return fmt.Sprintf("%s超时", e.Key)
}

type errNotFound struct {
	Key   string
	Value string
	Code  int
}

func ErrNotFound(res, name string) error {
	return &errNotFound{Key: res, Value: name, Code: ST_DATA_NOTFOUND}
}

func (e *errNotFound) Error() string {
	return fmt.Sprintf("%s[%s] not exists in es", e.Key, e.Value)
}

func (e *errNotFound) RetCode() int {
	return ST_DATA_NOTFOUND
}

func (e *errNotFound) Message() string {
	return fmt.Sprintf("%s不存在", e.Key)
}

type errSvcNotFound struct {
	Key   string
	Value string
	Code  int
}

func ErrSvcNotFound(svc, name string) error {
	return &errSvcNotFound{Key: svc, Value: name, Code: ST_SERVICE_NOTFOUND}
}

func (e *errSvcNotFound) Message() string {
	return "服务未开通"
}

func (e *errSvcNotFound) Error() string {
	return fmt.Sprintf("%s[%s] not exists in es", e.Key, e.Value)
}

func (e *errSvcNotFound) RetCode() int {
	return ST_SERVICE_NOTFOUND
}

////////
type errNoPermission struct {
	Resource string
	Action   string
	Who      string
	Code     int
}

func ErrNoPermission(who, resource, action string) error {
	return &errNoPermission{Resource: resource, Who: who, Action: action, Code: ST_AUTH_FAILURE}
}

func (e *errNoPermission) Error() string {
	return fmt.Sprintf("%s has no permission for %s:%s", e.Who, e.Resource, e.Action)
}

func (e *errNoPermission) RetCode() int {
	return ST_AUTH_FAILURE
}

func (e *errNoPermission) Message() string {
	return fmt.Sprintf("没有权限")
}

/*
	ST_OK            = 200
	ST_ARGS_ERROR    = 401
	ST_DATA_NOTFOUND = 404
	ST_SERVICE_NOTFOUND = 405
	ST_SER_ERROR     = 501
	ST_SER_BUSY      = 502
	ST_TOKEN_OUT   = 601
	ST_AUTH_FAILURE  = 701
*/

/////
type errToken struct {
	Code int
}

func ErrToken() error {
	return &errToken{Code: ST_TOKEN_OUT}
}

func (e *errToken) Error() string {
	return fmt.Sprintf("token error")
}

func (e *errToken) RetCode() int {
	return ST_TOKEN_OUT
}
func (e *errToken) Message() string {
	return fmt.Sprintf("token失效")
}

//////
type errInnerServer struct {
	Code int
	Func string
	Err  error
}

func ErrInnerServer(f string, err error) error {
	return &errInnerServer{Code: ST_SER_ERROR, Func: f, Err: err}
}

func (e *errInnerServer) Error() string {
	return fmt.Sprintf("%s err:%s", e.Func, e.Err.Error())
}

func (e *errInnerServer) RetCode() int {
	return ST_SER_ERROR
}

func (e *errInnerServer) Message() string {
	return fmt.Sprintf("服务器内部错误")
}

/////
func ErrorMsg(e error) (int, string, string) {
	te, ok := e.(TpaasError)
	if !ok {
		return ST_SER_ERROR, InnerServerErr, e.Error()
	}
	return te.RetCode(), te.Message(), te.Error()
}
