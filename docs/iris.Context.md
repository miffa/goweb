# Context 概述

## 参数解析
```go
	 ctx.Params().Get("id")   // get values from restful api  //bbbbbb:8080/xxx/{id}
	 ctx.Params().GetString("id")   // get values from restful api
	 ctx.Params().GetIntXXX("id")   // get values from restful api
	 ctx.Params().GetIntXXX("id")   // get values from restful api
	 ctx.Params().GetUintXXX("id")  // get values from restful api
	 ctx.Params().GetFloatXXX("id")  // get values from restful api

	 ctx.ReadJSON(&m)        // body is json string
	 ctx.FormValue("name")  // get value from the URL field's query parameters and the POST or PUT form data
	 ctx.UrlParamXXX("")        // get value from url query parameters
	 ctx.PostValueXXX("ident")  // get value from post put and PATCH body

```

## Context 解释


```go

package context
type Context interface {
// BeginRequest 针对每一个请求都会执行
// 它应该为新的请求准备(新的或从pool获得的)上下文的字段。

// 要跟随iris的流程，开发人员应:
// 1. 重置handler 为nil
// 2. 重置 values 为空
// 3. 重置会话为 nil
// 4. 重置 response writer 到 http.ResponseWriter
// 5. 重置 request 到 *http.Request
// 任何其他可选步骤都取决于开发的应用程序类型。
BeginRequest(http.ResponseWriter, *http.Request)

// 在请求被响应以后 执行EndRequest，并且这个请求的上下文变为无效或者已经释放
// 要跟随iris的流程，开发人员应:
// 1.刷新响应编写器的结果
// 2.刷新响应编写器的结果
// 任何其他可选步骤都取决于开发的应用程序类型。
EndRequest()

//ResponseWriter按预期返回与http.ResponseWriter兼容的响应编写器。
ResponseWriter() ResponseWriter

// 改变或者升级ResponseWriter.
ResetResponseWriter(ResponseWriter)

// 按预期返回原始的* http.Request。
Request() *http.Request

// SetCurrentRouteName在内部设置路由的名称，
//为了能够找到正确的当前“只读”路由时
// end-developer调用`GetCurrentRoute（）`函数。
//如果您更改了该名称，它将由路由器初始化,他只会更改当名称
// 通过`GetCurrentRoute（）`您将会得到所更改的名称。

//相反，要执行不同的路径
//你应该使用`Exec`函数
//或通过`SetHandlers 或者 AddHandler`函数更改处理程序。
SetCurrentRouteName(currentRouteName string)

// GetCurrentRoute返回当前注册的“只读”路由
//正在注册此请求的路径。
//r:=ctx.GetCurrentRoute()//获取当前路由
//fmt.Println(r)//输出:GET /users/{id:int}/profile
GetCurrentRoute() RouteReadOnly

//调用  SetHandlers(handlers)
// 并且执行该方法所设置的第一个handlers
// handlers 不应该为空.
// 这是路由器使用的 开发者应该使用exex(Handler)直接调用要执行的handler
Do(Handlers)

// AddHandler可以添加处理程序
//到服务时间的当前请求，
//这些处理程序没有持久化到路由器。
//路由器正在调用此函数来添加路由的处理程序。
//如果调用了AddHandler，则会插入处理程序
//到已经定义的路由处理程序的末尾。
//ctx.Handlers()[1](ctx)//调用当前添加的函数1为当前handler再handlers数组中的下标
//再当前被添加的函数中使用ctx.HandlerIndex(2)定义该函数的位置
AddHandler(...Handler)

//SetHandlers用新的 handler取代所有处理程序。
SetHandlers(Handlers)

//返回当前所有的可调用的handler。
Handlers() Handlers

// 设置当前hander 在handeler数组中的下标位置返回-1代表ok
// Look Handlers(), Next() and StopExecution() too.
HandlerIndex(n int) (currentIndex int)

// Proceed 是检查特定处理程序是否已被执行并在其中调用`ctx.Next`函数的另一种方法
// 只有在内部运行处理程序时，这才有用
// 另一个处理它只是在索引和after索引之前检查。

// 一个使用实例就是去执行一个中间件
// 在控制器的`BeginRequest`中调用其中的`ctx.Next`。
// Controller查看整个流程（BeginRequest，方法处理程序，EndRequest） 作为一个处理程序，
// 所以`ctx.Next`不会反映到方法处理程序中
// 如果从`BeginRequest`调用。
//
// 虽然`BeginRequest`不应该用于调用其他处理程序，
//引入了`BeginRequest`以便能够设置 执行前所有方法处理程序的公共数据。
// 控制器可以正常接受来自MVC的应用程序路由器的中间件。
// 让我们看一个`ctx.Proceed`的例子：
// var authMiddleware = basicauth.New(basicauth.Config{
//         Users: map[string]string{
//             "admin": "password",
//         },
// })
// func (c *UsersController) BeginRequest(ctx iris.Context) {
//     if !ctx.Proceed(authMiddleware) {
//         ctx.StopExecution()
//      }
// }
//这个Get（）将在与`BeginRequest`相同的处理程序中执行，
//内部控制器检查`ctx.StopExecution`。
//因此，如果BeginRequest调用了`StopExecution`，它将不会被触发。
// func(c *UsersController) Get() []models.User {
//      return c.Service.GetAll()
// }
// 另一种方法是`！ctx.IsStopped（）`如果中间件在失败时使用`ctx.StopExecution（）`。
Proceed(Handler) bool

// HandlerName返回当前处理程序的名称，有助于调试。格式package.function
HandlerName() string

/ Next从处理程序链中调用所有下一个处理程序,
//它应该在中间件中使用
//注意：自定义上下文应该重写此方法，以便能够传递自己的context.Context实现。
Next()
// NextOr检查链是否有下一个处理程序，如果是，则执行它
//否则它根据给定的处理程序设置分配给此Context的新链
//并执行其第一个处理程序。

//如果存在并执行下一个处理程序，则返回true，否则返回false
//请注意，如果找不到下一个处理程序，则缺少处理程序
//它向客户端发送未找到状态（404）并停止执行。
NextOr(handlers ...Handler) bool

// NextOrNotFound检查链是否有下一个处理程序，如果是，则执行它
//否则它会向客户端发送未找到状态（404）并停止执行。
//如果存在并执行下一个处理程序，则返回true，否则返回false
NextOrNotFound() bool

// NextHandler从处理程序链返回（它不执行）下一个处理程序。
//如果需要执行下一个返回处理程序，请使用.Skip（）跳过此处理程序。
NextHandler() Handler
//跳过/忽略处理程序链中的下一个处理程序，
//它应该在中间件中使用
Skip()
//终止当前程序但不是退出（等于设定一个终止标记）
//如果调用StopExecution，则调用以下.Next调用被忽略，
//因此链中的下一个处理程序不会被触发。
StopExecution()

// IsStopped检查并在Context的当前位置为255时返回true，
//表示调用了StopExecution（）
IsStopped() bool

// 请求参数介绍
// Params返回当前url的命名参数键值存储。
//这里保存了命名路径参数。
//作为整个Context，此存储是按请求生存期。
Params() *RequestParams

/值返回当前的“用户”存储。
//可以在此处保存命名路径参数和任何可选数据。
//作为整个Context，此存储是按请求生存期。
//您可以使用此函数来设置和获取本地值
//可用于在处理程序和中间件之间共享信息。
Values() *memstore.Store

//翻译是i18n（本地化）中间件的功能，
//它调用Get（“translate”）来返回翻译的值。
//示例：https：//github.com/kataras/iris/tree/master/_examples/miscellaneous/i18n
Translate(format string, args ...interface{}) string
//  +------------------------------------------------------------+
//  |路径,主机,子域,IP,标头等...              |
//  +------------------------------------------------------------+
// 返回request.Method，客户端的http方法返回给服务器
Method() string

// Path返回完整的请求路径，/name 不包含host 和请求参数
//如果EnablePathEscape配置字段为true，则进行转义。 
Path() string

// RequestPath返回完整的请求路径
//基于'escape'。
RequestPath(escape bool) string

// Host返回当前url的主机部分。如localhost
Host() string

// Subdomain返回此请求的子域（如果有）。
//请注意，这是一种不能涵盖所有情况的快速方法。
Subdomain() (subdomain string)

//如果当前子域（如果有）是www，则IsWWW返回true。
IsWWW() bool

/ RemoteAddr尝试解析并返回真实客户端的请求IP。
//基于可从Configuration.RemoteAddrHeaders修改的允许标头名称。
//如果基于这些头的解析失败，那么它将返回Request的`RemoteAddr`字段
//在HTTP处理程序之前由服务器填充。

//查看  `Configuration.WithRemoteAddrHeader(...)`,
//      `Configuration.WithoutRemoteAddrHeader(...)` for more.
RemoteAddr() string

// GetHeader根据名称返回请求标头的值。
GetHeader(name string) string

//如果此请求是'ajax请求'（XMLHttpRequest），则IsAjax返回true
//没有100％的方式知道请求是通过Ajax进行的。
//您永远不应该信任来自客户端的数据，可以通过欺骗轻松克服这些数据。

//注意“X-Requested-With”标题可以被任何客户端修改（因为“X-”），
//所以不要依赖IsAjax来做真正严肃的事情，
//尝试找到另一种检测类型的方法（即内容类型），
//有很多博客描述这些问题并提供不同类型的解决方案，
//它始终取决于您正在构建的应用程序，
//这就是为什么这个'IsAjax``足够简单以便通用的原因。
//阅读更多信息：
IsAjax() bool

// IsMobile会检查客户端是否正在使用移动设备（手机或平板电脑）与此服务器通信。
//如果返回值为true，则表示http客户端使用移动设备
//设备与服务器通信，否则为false。
//请注意，这会检查“User-Agent”请求标头。
IsMobile() bool
//  +------------------------------------------------------------+
//  |响应头助手                                   |
//  +------------------------------------------------------------+
// 添加一个响应头
Header(name string, value string)

//设置响应头 "Content-Type" to the 'cType'.
ContentType(cType string)

// GetContentType返会响应头 “Content-Type”
//可以使用'ContentType'设置之前。
GetContentType() string

// GetContentLength 返回请求头 "Content-Length". 的值
//当改值没有找到或者不是一个数字 将会返回0
GetContentLength() int64

// StatusCode 设置响应的statu code 值

StatusCode(statusCode int)

//返回响应的当前状态代码。
GetStatusCode() int

//重定向向客户端发送重定向响应
//到特定网址或相对路径。
//接受2个参数字符串和一个可选的int
//第一个参数是要重定向的URL
//第二个参数是应该发送的http状态，
//默认为302（StatusFound），
//你可以将它设置为301（Permant重定向）
//或303（StatusSeeOther）如果POST方法，
//或StatusTemporaryRedirect（307），如果这是nessecery。
Redirect(urlToRedirect string, statusHeader ...int)
//  +------------------------------------------------------------+
//  | 各种请求和Post数据                            |
//  +------------------------------------------------------------+
// URLParam 如果url参数存在，则返回true，否则返回false。
URLParamExists(name string) bool

// URLParamDefault返回请求中的get参数，
//如果没有找到 则返回 设定的 "def".
URLParamDefault(name string, def string) string

//URLParam返回请求中的get参数（如果有）。
URLParam(name string) string

// 返回url查询参数，其中从请求中删除了尾随空格。
URLParamTrim(name string) string

// URLParamTrim 从请求中返回转义的url查询参数。URLParamEscape（名称字符串）字符串
URLParamEscape(name string) string

// URLParamInt将url查询参数作为来自请求的int值返回，
//返回-1，如果解析失败则返回错误。
URLParamInt(name string) (int, error)

//将url查询参数作为来自请求的int值返回，
//如果找不到或解析失败，则返回 设定的 "def".
URLParamIntDefault(name string, def int) int

// URLParamInt64 将url查询参数作为来自请求的int64值返回,
//返回-1，如果解析失败则返回错误。
URLParamInt64(name string) (int64, error)

// URLParamInt64Default  将url查询参数作为来自请求的int64值返回,
//如果找不到或解析失败，则返回 设定的 "def".
URLParamInt64Default(name string, def int64) int64

// URLParamFloat64 将url查询参数作为来自请求的float64值返回,
//返回-1，如果解析失败则返回错误。
URLParamFloat64(name string) (float64, error)

// URLParamFloat64Default  将url查询参数作为来自请求的float64值返回,
//如果找不到或解析失败，则返回 设定的 "def".
URLParamFloat64Default(name string, def float64) float64
// URLParamFloat64 将url查询参数作为来自请求的bool值返回,
//返回-1，如果解析失败则返回错误。

URLParamBool(name string) (bool, error)

// URLParams返回由逗号分隔的GET查询参数的映射（如果有多个）
//如果找不到任何内容，则返回空map。

URLParams() map[string]string

// FormValueDefault通过其"名称"返回单个解析的表单值,
//包括URL字段的查询参数和POST或PUT表单数据。
//如果没有找到返回设定的def
FormValueDefault(name string, def string) string

// FormValue通过其“名称”返回单个解析的表单值，
//包括URL字段的查询参数和POST或PUT表单数据。
FormValue(name string) string

// FormValues返回已解析的表单数据，包括URL
//字段的查询参数和POST或PUT表单数据。

//默认表单的内存最大大小为32MB，可以通过更改
//在主要配置中的`iris＃WithPostMaxMemory`配置器传递给`app.Run`的第二个参数。

//注意：需要检查nil。
FormValues() map[string][]string

// PostValueDefault从POST，PATCH返回解析后的表单数据
//或基于“名称”的PUT身体参数。
//如果没有找到返回设定的def
// If not found then "def" is returned instead.
PostValueDefault(name string, def string) string

// PostValue从POST，PATCH返回解析后的表单数据
//或基于“名称”的PUT身体参数
PostValue(name string) string

// PostValueTrim从POST，PATCH返回解析后的表单数据
//或PUT基于“名称”的体参数，没有尾随空格。
PostValueTrim(name string) string

// PostValueInt从POST，PATCH返回解析后的表单数据
//或PUT基于“名称”的体参数，如int。

//如果未找到则返回-1并返回非nil错误。
PostValueInt(name string) (int, error)

// PostValueIntDefault从POST，PATCH返回解析后的表单数据
//或PUT基于“名称”的体参数，如int

//如果没有找到返回设定的def
PostValueIntDefault(name string, def int) int

// PostValueInt64从POST，PATCH返回解析后的表单数据，
//或PUT基于“名称”的体参数，如float64。
// 如果未找到则返回-1并返回非nil错误。
PostValueInt64(name string) (int64, error)

// PostValueInt64Default从POST，PATCH返回解析后的表单数据，
//或PUT基于“名称”的体参数，如int64。
//如果没有找到返回设定的def
PostValueInt64Default(name string, def int64) int64

// PostValueFloat64从POST，PATCH返回解析后的表单数据，
//或PUT基于“名称”的体参数，如float64。
// 如果未找到则返回-1并返回非nil错误。
PostValueFloat64(name string) (float64, error)
// PostValueInt64Default从POST，PATCH返回解析后的表单数据，
//或PUT基于“名称”的体参数，如Int64。
//如果没有找到返回设定的def
PostValueFloat64Default(name string, def float64) float64

// PostValueInt64Default从POST，PATCH返回解析后的表单数据，
//或PUT基于“名称”的身体参数，如bool。

//如果未找到或value为false，则返回false，否则返回true。
PostValueBool(name string) (bool, error)

// PostValues从POST，PATCH返回所有已解析的表单数据，
//或PUT体参数基于“名称”作为字符串切片。
//
//默认表单的内存最大大小为32MB，可以通过更改
//在主要配置中的`iris＃WithPostMaxMemory`配置器传递给`app.Run`的第二个参数。

PostValues(name string) []string

// FormFile返回从客户端收到的第一个上传文件。
//默认表单的内存最大大小为32MB，可以通过更改
//在主要配置中的`iris＃WithPostMaxMemory`配置器传递给`app.Run`的第二个参数。
//示例：https：//github.com/kataras/iris/tree/master/_examples/http_request/upload-file

FormFile(key string) (multipart.File, *multipart.FileHeader, error)
// UploadFormFiles从客户端上传任何收到的文件
//到系统物理位置“destDirectory”。

//第二个可选参数“before”为调用者提供了机会
//在保存到磁盘之前修改* miltipart.FileHeader，
//它可以用来根据当前请求更改文件名，
//可以更改所有FileHeader的选项。你可以忽略它
//在将文件保存到磁盘之前，您不需要使用此功能。
//请注意，它不会检查请求正文是否已流式传输。

//将复制的长度返回为int64和
//如果至少有一个新文件，则为非nil错误
//由于操作系统的权限或无法创建//
// http.ErrMissingFile如果没有收到文件。

//如果您想要接收和接受文件并手动管理它们，您可以使用`context＃FormFile`
//而是创建一个适合您需要的复制功能，下面是一般用法。

//默认表单的内存最大大小为32MB，可以通过更改
//在主要配置中的`iris＃WithPostMaxMemory`配置器传递给`app.Run`的第二个参数。

//参见`FormFile`更受控制以接收文件。
//示例：https：//github.com/kataras/iris/tree/master/_examples/http_request/upload-files
UploadFormFiles(destDirectory string, before ...func(Context, *multipart.FileHeader)) (n int64, err error)
//  +------------------------------------------------------------+
//  | 自定义HTTP错误                                               |
//  +------------------------------------------------------------+
// NotFound使用特定的自定义错误错误处理程序向客户端发出错误404。
//请注意，如果您不想使用下一个处理程序，则可能需要调用ctx.StopExecution（）
//被执行下一个处理程序正在iris上执行，因为你可以使用它
//错误代码并将其更改为更具体的错误代码，即
// users：= app.Party（“/ users”）
// users.Done（func（ctx context.Context）{if ctx.StatusCode（）== 400 {/ * / users * /}的自定义错误代码}）
NotFound()
//  +------------------------------------------------------------+
//  | Body Readers                                               |
//  +------------------------------------------------------------+
// SetMaxRequestBodySize设置请求主体大小的限制
//应该在从客户端读取请求主体之前调用。
SetMaxRequestBodySize(limitOverBytes int64)

// UnmarshalBody读取请求的正文并将其绑定到任何类型的值或指针。
//用法示例：context.ReadJSON，context.ReadXML。

//例如: https://github.com/kataras/iris/blob/master/_examples/http_request/read-custom-via-unmarshaler/main.go
UnmarshalBody(outPtr interface{}, unmarshaler Unmarshaler) error
// ReadJSON从请求的主体读取JSON，并将其绑定到任何json-valid类型的值的指针。

// 例如: https://github.com/kataras/iris/blob/master/_examples/http_request/read-json/main.go
ReadJSON(jsonObjectPtr interface{}) error
// ReadXML从请求的正文中读取XML，并将其绑定到任何xml-valid类型值的指针。

// 例如: https://github.com/kataras/iris/blob/master/_examples/http_request/read-xml/main.go
ReadXML(xmlObjectPtr interface{}) error

// ReadForm将formObject与表单数据绑定在一起
//它支持任何类型的结构。
// Example: https://github.com/kataras/iris/blob/master/_examples/http_request/read-form/main.go
ReadForm(formObjectPtr interface{}) error
//  +------------------------------------------------------------+
//  | Body (raw) Writers                                         |
//  +------------------------------------------------------------+
// Write将数据作为HTTP回复的一部分写入连接。
//如果尚未调用WriteHeader，则写入调用
//在写入数据之前写入WriteHeader（http.StatusOK）。如果是标题
//不包含Content-Type行，Write添加Content-Type集
//将最初的512字节写入数据传递给的结果
// DetectContentType。

//根据HTTP协议版本和客户端，调用
// Write或WriteHeader可能会阻止将来读取
// Request.Body。对于HTTP / 1.x请求，处理程序应该读取任何内容
//在编写响应之前需要请求正文数据。一旦
//标头已被刷新（由于显式的Flusher.Flush
//调用或写入足够的数据来触发刷新），即请求体
//可能无法使用对于HTTP / 2请求，Go HTTP服务器允许
//处理程序在同时继续读取请求体
//写回复但是，可能不支持此类行为
//由所有HTTP / 2客户端。处理者应在写作前阅读
//可以最大化兼容性。
Write(body []byte) (int, error)

// Writef根据格式说明符格式化并写入响应。

//返回写入的字节数和遇到的任何写入错误。
Writef(format string, args ...interface{}) (int, error)

// WriteString将一个简单的字符串写入响应。

//返回写入的字节数和遇到的任何写入错误。
WriteString(body string) (int, error)

// SetLastModified根据“modtime”输入设置“Last-Modified”。
//如果“modtime”为零，那么它什么都不做。

//它主要内部在核心/路由器和上下文包上。
//注意正在使用modtime.UTC（）而不仅仅是modtime，所以
//你不必知道内部结构才能使其有效。
SetLastModified(modtime time.Time)
// CheckIfModifiedSince检查自“modtime”以来是否修改了响应。
//请注意，它与服务器端缓存无关。
//它通过检查“If-Modified-Since”请求标头来执行这些检查
//由客户端或以前的服务器响应头发送
//（例如WriteWithExpiration或StaticEmbedded或Favicon等）
//是有效的，它在“modtime”之前。

//检查！modtime && err == nil是必要的，以确保
//它没有被修改，因为它可能会返回false而不是偶数
//由于某些错误，有机会检查客户端（请求）标头，
//喜欢HTTP方法不是“GET”或“HEAD”或者“modtime”是零
//或者从头部解析时间失败。

//它主要用于内部，例如`环境＃WriteWithExpiration`。
//注意正在使用modtime.UTC（）而不仅仅是modtime，所以
//你不必知道内部结构才能使其有效。
CheckIfModifiedSince(modtime time.Time) (bool, error)

// WriteNotModified向客户端发送304“未修改”状态代码，
//它确保内容类型，内容长度标题
//在发送响应之前删除任何“ETag”。

//它主要在core / router / fs.go和context方法的内部使用。
WriteNotModified()

// WriteWithExpiration类似于Write但它发送的是到期日期时间
//刷新每个包级别的“StaticCacheDuration”字段。
WriteWithExpiration(body []byte, modtime time.Time) (int, error)

// StreamWriter注册给定的流编写器以进行填充
//回复正文

//禁止从作者访问上下文和/或其成员。
//此功能可用于以下情况：

// *如果响应主体太大（超过iris.LimitRequestBodySize（如果已设置））。
// *如果响应正文从缓慢的外部源流式传输。
// *如果必须以块的形式将响应主体流式传输到客户端。
//（又名`http server push`）。

//接收一个接收响应编写器的函数
//并在它应该停止写入时返回false，否则为true以便继续
StreamWriter(writer func(w io.Writer) bool)
//  +------------------------------------------------------------+
//  | Body Writers with compression                              |
//  +------------------------------------------------------------+
//如果客户端支持gzip压缩，则ClientSupportsGzip重新为true。
ClientSupportsGzip() bool

// WriteGzip接受字节，压缩为gzip格式并发送到客户端。
//返回写入的字节数和错误（如果客户端不支持gzip压缩）
//您可以在同一个处理程序中重用此函数
//多次写入更多数据，没有任何麻烦
WriteGzip(b []byte) (int, error)
// TryWriteGzip接受字节，压缩为gzip格式并发送到客户端。
//如果客户端不支持gzip，则内容按原样写入，未压缩。
TryWriteGzip(b []byte) (int, error)

// GzipResponseWriter将当前响应编写器转换为响应编写器
//当它的.Write调用它时，将数据压缩为gzip并将它们写入客户端。
//
//也可以使用.Disable和.ResetBody来禁用它以回滚到通常的响应编写器。
GzipResponseWriter() *GzipResponseWriter

//如果客户端，Gzip启用或禁用（如果在之前启用）gzip响应编写器
//支持gzip压缩，因此以下响应数据将会
//作为压缩的gzip数据发送到客户端。
Gzip(enable bool)
//  +------------------------------------------------------------+
//  | Rich Body Content Writers/Renderers                        |
//  +------------------------------------------------------------+
// ViewLayout在.View时设置“layout”选项
//后来在同一个请求中被调用。
//当需要根据链中的先前处理程序设置或/和更改布局时很有用。
//
//注意'layoutTmplFile'参数可以设置为iris.NoLayout || view.NoLayout
//禁用特定视图渲染操作的布局，
//它会禁用引擎配置的布局属性。

// Look .ViewData and .View too.

// Example: https://github.com/kataras/iris/tree/master/_examples/view/context-view-data/
ViewLayout(layoutTmplFile string)

// ViewData保存一个或多个键值对，以便在.View时传递
//后来在同一个请求中被调用。
//当需要设置或/和更改链中先前hanadler的模板数据时很有用。

//如果.View的“绑定”参数不是nil而且它不是一种地图
//然后忽略这些数据，绑定具有优先级，因此主路径的处理程序仍然可以决定。
//如果绑定是map或context.Map，那么这些数据将被添加到视图数据中
//并传递给模板。

//在.View之后，数据不会被销毁，以便在需要时重新使用（同样，在与其他所有相同的请求中），
//清除视图数据，开发人员可以调用：
// ctx.Set（ctx.Application（）。ConfigurationReadOnly（）。GetViewDataContextKey（），nil）

//如果'key'为空，则将值添加为（struct或map），并且开发人员无法添加其他值。

// Look .ViewLayout and .View too.
// Example: https://github.com/kataras/iris/tree/master/_examples/view/context-view-data/
ViewData(key string, value interface{})

// GetViewData返回`context #ViewData`注册的值。
//返回值是`map [string] interface {}`，这意味着
//如果一个自定义结构注册到ViewData那么这个函数
//将尝试解析它以进行映射，如果失败则返回值为nil
//如果不同，检查零是一个很好的做法
//通过`ViewData`注册了一些值或没有数据。
//类似于这样 `viewData := ctx.Values().Get("iris.viewData")` 或者
//`viewData := ctx.Values().Get(ctx.Application().ConfigurationReadOnly().GetViewDataContextKey())`.
GetViewData() map[string]interface{}
// View根据注册的视图引擎呈现模板。
//第一个参数接受相对于视图引擎的目录和扩展名的文件名，
//即：如果目录是“./templates”并想要呈现“./templates/users/index.html”
//然后你传递“users / index.html”作为文件名参数。

//第二个可选参数可以接收单个“视图模型”
//如果它不是nil，它将绑定到视图模板，
//否则会检查`ViewData`存储的先前视图数据
//即使存储在任何先前的handlerv（中间件）中，也存在相同的请求。

//也看.ViewData`和.ViewLayout。

//示例：https://github.com/kataras/iris/tree/master/_examples/view
View(filename string, optionalViewModel ...interface{}) error

// 二进制将原始字节写为二进制数据。
Binary(data []byte) (int, error)
// Text将字符串写为纯文本。
Text(text string) (int, error)
// HTML将字符串写为 text/html。
HTML(htmlContents string) (int, error)
// JSON 整理给定的接口对象并写入JSON响应。
JSON(v interface{}, options ...JSON) (int, error)
// JSONP  整理给定的接口对象并写入JSON响应
JSONP(v interface{}, options ...JSONP) (int, error)
// XML  整理给定的接口对象并写入xml响应.
XML(v interface{}, options ...XML) (int, error)
//Markdown 将markdown解析为html并将其结果呈现给客户端。
Markdown(markdownB []byte, options ...Markdown) (int, error)
// 用yaml解析器解析“v”并将其结果呈现给客户端。
YAML(v interface{}) (int, error)
//  +------------------------------------------------------------+
//  | Serve files                                                |
//  +------------------------------------------------------------+
// ServeContent提供内容，标题是自动设置的
//接收三个参数，它是低级函数，而不是你可以使用.ServeFile（string，bool）/ SendFile（string，string）

//在此函数调用之前，您可以使用`context＃ContentType`定义自己的“Content-Type”。

//此函数不支持恢复（按范围），
//使用ctx.SendFile或路由器的`StaticWeb`代替。
ServeContent(content io.ReadSeeker, filename string, modtime time.Time, gzipCompression bool) error

// ServeFile提供一个文件（例如发送一个文件，一个zip到客户端你应该使用`SendFile`）
//接收两个参数
//文件名/路径（字符串）
// gzipCompression（bool）

//在此函数调用之前，您可以使用`context＃ContentType`定义自己的“Content-Type”。
//
//此函数不支持恢复（按范围），
//使用ctx.SendFile或路由器的`StaticWeb`代替。

//当您想要向客户端提供动态文件时使用它。
ServeFile(filename string, gzipCompression bool) error

// SendFile将文件强制下载发送到客户端
//使用此代替ServeFile来“强制下载”更大的文件到客户端。
SendFile(filename string, destinationName string) error
//  +------------------------------------------------------------+
//  | Cookies                                                    |
//  +------------------------------------------------------------+
// SetCookie添加了一个cookie
SetCookie(cookie *http.Cookie)

// SetCookieKV添加一个cookie，只接收一个名字（字符串）和一个值（字符串）

//如果您使用此方法，它将在2小时后到期
//如果要更改更多字段，请使用ctx.SetCookie或http.SetCookie。
SetCookieKV(name, value string)

// GetCookie按名称返回cookie的值
//如果没有找到，则返回空字符串。
GetCookie(name string) string
// RemoveCookie按名称删除cookie。
RemoveCookie(name string)

// VisitAllCookies接受一个循环的访问者
//在每个（请求的）cookie的名称和值上。
VisitAllCookies(visitor func(name string, value string))

// MaxAge返回“缓存控制”请求标头的值
//秒为int64
//如果未找到标头或解析失败，则返回-1。
MaxAge() int64
//  +------------------------------------------------------------+
//  | Advanced: Response Recorder and Transactions               |
//  +------------------------------------------------------------+
// Record将上下文的基本和直接responseWriter转换为ResponseRecorder
//可用于重置body，重置标头，获取body，
//随时随地获取和设置状态代码。
Record()

// Recorder返回上下文的ResponseRecorder
//如果没有录制，则开始录制并返回新上下文的ResponseRecorder
Recorder() *ResponseRecorder

// IsRecording返回响应记录器和真值
//当响应作者正在记录状态代码，正文，标题等时，
// else返回nil和false。
IsRecording() (*ResponseRecorder, bool)

/ BeginTransaction启动范围内的事务。
//您可以搜索有关业务交易如何运作的第三方文章或书籍（这很简单，尤其是在这里）。
//请注意，这是独一无二的
//（=我在Golang上从未见过关于此主题的任何其他示例或代码，到目前为止，与大多数虹膜功能一样......）
//它不包括所有路径
//例如数据库，这应该由用于建立数据库连接的库来管理，
//此事务范围仅用于上下文的响应。
//交易也有自己的中间件生态系统，看看iris.go：UseTransaction。

// 参考 https://github.com/kataras/iris/tree/master/_examples/ for more
BeginTransaction(pipe func(t *Transaction))

//如果调用SkipTransactions，则跳过其余的事务
//如果在第一次交易之前调用，则为全部
SkipTransactions()
//如果事务被跳过或取消，则TransactionsSkipped返回true。
TransactionsSkipped() bool

// Exec调用`context / Application #ServeCtx`
//基于此上下文但具有更改的方法和路径
//喜欢它是用户请求的，但事实并非如此。
//
//离线表示路由已注册到虹膜，并具有正常路由所具有的所有功能
//但它不能通过浏览获得，它的处理程序仅在其他处理程序的上下文调用它们时执行
//它可以验证路径，有会话，路径参数等等。

//你可以通过app.GetRoute找到路线（“theRouteName”）
//您可以将路径名称设置为：myRoute：= app.Get（“/ mypath”，handler）（“theRouteName”）
//将为路由设置名称并返回其RouteInfo实例以供进一步使用。

//它不会更改全局状态，如果路由处于“离线状态”，它将保持脱机状态。
// app.None（...）和app.GetRoutes（）。离线（路线）/。在线（路线，方法）
//示例：https：//github.com/kataras/iris/tree/master/_examples/routing/route-state

//用户可以通过简单的使用rec：= ctx.Recorder（）获得响应; rec.Body（）/ rec.StatusCode（）/ rec.Header（）。
//保留Context的值和Session，以便能够通过结果路由进行通信。

//这是针对极端用例的，99％的情况永远不会对你有用。
Exec(method, path string)

// RouteExists报告特定路由是否存在
//如果不在根域内 ，它将从上下文主机的当前子域搜索。
RouteExists(method, path string) bool

// Application返回属于此上下文的iris app实例。
//值得注意的是这个函数返回一个接口
//应用程序，包含安全的方法
//在服务时间执行。完整的应用程序的字段
//这里没有方法可供开发人员使用。
Application() Application

// String返回此请求的字符串表示形式。
//每个上下文都有一个唯一的字符串表示
//它可以用于简单的调试场景，即打印上下文为字符串。
//它返回什么？一个声明长度的数字
//跟随每个可执行应用程序的总`String`调用
//通过远程IP（客户端），最后是方法：url。
String() string
}

```


