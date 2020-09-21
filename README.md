## 初始化工程
+ ./initproject.sh  your_name

+ 文件以及目录介绍
  +  build.sh      : build脚本      ./build.sh  myversion    
  +  CHANGELOG.md  : 版本的changelog（选用）
  +  cmd/mybinary  ：程序入口 main.go
  +  conf          : 配置文件，采用yaml格式 
  +  docs          ：程序说明文档（选用）
  +  go.mod     
  +  go.sum  
  +  initproject.sh: 工程初始化脚本，目前仅支持linux和macos
  +  Makefile        
  +  pkg           ：业务代码目录 
  +  README.md     
  +  start.sh      ：启动脚本demo，自己定制
  +  vendor        ：vendor目录

## pkg目录
以下是参考目录，可以按照自己爱好自行定义  
+ config    ：配置文件读写类
+ define    ：各个模块公共使用的一些常量、错误码和错误类定义，目前封装了返给controller（handler）的常见错误（参数错误、权限错误、token错误、内部错误、服务未开通错误）
+ router    ：路由注册文件夹， router.go文件配置所有的路由信息
+ middleware：中间件处理类，token校验、审计日志等公共操作都可以放到这里处理
+ handler   ：controller处理层，处理
+ service   ：业务处理层
+ dao       : 数据操作
+ models    ：对象定义文件夹
+ thirdurl  ：第三方服务访问客户端（可选）
+ iocgo     ：单件类或者全局变量资源管理类（帮开发人员初始化和销毁单件类）
+ libs      ：模块公用的lib库
+ version   ：应用的名称和版本信息

## 使用手册
### 路由url命名格式统一为 
>    /服务名称/api/版本/资源key/资源value/子资源key/子资源value...  

#### 平台级别
     eg: http://mock.com/{service}/api/v1/resources/{resource}/clusters/{cluster}/...

###  部门or项目级别资源操作
     eg：http://mock.com/{service}/api/v1/apps/{app_id}/clusters/{cluster name}/...
     eg：http://mock.com/{service}/api/v1/namespace/{namespace_id}/clusters/{cluster name}/...

### 中间件注册&路由注册
iris/pkg/router/router.go
```
  //中间件注册
  a.SetMiddleware(handler.RequestLog)
  //路由注册
  a.Post("/demopost", handler.Demo2)
  
  // 子路由&中间件
  p := a.Group("/swordmen_novel", middleware.DemoPartyMiddleware)
  // p.Use(middleware.DemoPartymiddleware2)  // 方式2
  p.Get("/gulong", handler.Demo) 
```

### 请求处理
handler package
```
 func Demo3(ctx iris.Context) {
     retdata, err := service.GetSingleTon().Demook3() 
     if err != nil {
         ResponseErrMsg(ctx, err)
         return
     }
     ReponseOk(ctx, retdata)
     log.Debugf("demo reponse ok")
 }
```

### 参数解析
  更多参见docs/resful.md  
```
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
### 应答处理  
+ handler中的函数返会给前端时统一调用一下处理函数
  +  handler.ReponseOk       返回正确应答消息
  +  handler.ResponseErr     返回错误应答消息（可以自由定义错误码和错误消息，一般是controller层自己的错误使用此应答函数）
  +  handler.ResponseErrMsg  返回错误应答消息（使用的define.TpaasError封装错误消息，需要根据自己需求实现自己的TpaasError，一般service层返回的错误使用此应答函数）
+ 应答结构体  
```
 type RespJson struct {
     Status int         `json:"code"`     // 返给前端的业务码， 可以使用define中定义的业务码
     Msg    string      `json:"msg"`      // 返给前端的展示消息（一般用于出错时展示）
     Data   interface{} `json:"data"`     // 返给前端的数据
     Detail string      `json:detail, omitempty`  //返给前端的出错专业信息（前端不显示，方便开发人员排错）
 } 
```  
+ 应答码和应答错误  
>  应答错误码和定制的应答错误类都封装在define包中，可以按需定制（里面也附注了常用的http应答码，可以提供参考）

### 错误处理  
+  service层返给handler的错误尽量使用TpaasError类型封装一下处理，返给controller层洁
+  dao层或者其他模块返给service层的错误都是用error
+  如果是该模块自己的错误使用github.com/pkg/errors.New()生成或者提前定义好该模块常用错误
+  如果是调用第三方或者别的函数返回的错误要继续返给service，请使用github.com/pkg/errors.WithXXX函数把错误封装一下再返回给service
+  not found错误处理：把数据notfound当作一种error来处理已经是golang目前流行的一种做法，但是这种错误对业务使用者来讲造成了很不好的体验，所以在我们代码中尽量将此类的错误转义一下，转成我们业务需要的场景数据（一个空指针、一个默认对象、一个空数组等等）

### 日志
+ github.com/sirupsen/logrus
+ 日志以json格式输出
+ 请求和应答日志目前都已经封装记录到日志中
+ 日志内容组成
```  
    文件                    函数        日志级别            日志内容                            时间
{"file":"main.go","func":"main.main","level":"info","msg":"http pprof service init ok","time":"2020-06-04T10:11:09+08:00"}
```
+ 日志打印位置  
  + 统一打印到service层；
  + 如果业务足够简单没有service就打印到controller层
+ 日志内容
  + 调试型日志（debug）：
      做什么 参数是什么 期望什么 
      做什么 参数是什么 结果是什么
  + 记录型日志：
      做什么 参数是什么 结果是什么
  + 错误日志：
      做什么 参数是什么 错误是什么
  + 第三方服务调用日志
      调用什么 参数什么  结果是什么  耗时
+ 需要记录日志的场景
  + 问题排查
  + 外部调用跟踪
  + 数据状态变化
  + 函数的入口和出口
  + 异常错误
  + 非预期执行 
  + 关键步骤
  + 慢操作日志（第三方调用）

+ 日志必打
  + service层的非get请求，尽量都打印一条记录型日志
  + service层错误日志
  + 没有service层，在controller层打印错误日志
  + 数据删除和添加类型的操作
+ 注意事项
  + 日志内容要是可见字符，数据要有结构
  + 如果参数字段太多，打印关键参数
  + 不要打印重复的无意义的日志  ----  。。。。。 ******   111111111
  + 不要打印混视听的日志

### 全局资源or单件类管理
+ iocgo使用方法  

```
    import(
      "iris/pkg/libs/tokenbucket"
      "iris/pkg/config"
      "iris/pkg/iocgo"
    )
    
    type Ratelimiter struct {
    	*tokenbucket.Bucket
    }

    var RL *Ratelimiter
    var rlonce sync.Once

    func Ralimit() *Ratelimiter {
    	rlonce.Do(func() { RL = new(Ratelimiter) })
    	return RL
    }

    //！！！
    func (rl *Ratelimiter) Init(cfg *config.TpaasConfig) error {
    	return rl.InitFrq(1000)
    }
    //！！！
    func (rl *Ratelimiter) Close() error {
    	return rl.Bucket.Close()
    }
    
    // 注册到iocgo 帮助统一初始化和销毁
    func init() {
    	iocgo.Register("ratelimit_1000", Ralimit())
    }
    
    //////////////////////////////////////////////
    // 全局变量资源或者单件类资源初始化interface
     type Initializer interface {
         Init(cfg *config.TpaasConfig) error
         Close() error
    }

```

+ 
    

