# 中间件使用

## 示例
```go
    package main
    import "github.com/kataras/iris"
    func main() {
        app := iris.New()
        app.Get("/", before, mainHandler, after)
        app.Run(iris.Addr(":8080"))
    }
    func before(ctx iris.Context) {
        shareInformation := "this is a sharable information between handlers"

        requestPath := ctx.Path()
        println("Before the mainHandler: " + requestPath)

        ctx.Values().Set("info", shareInformation)
        ctx.Next() //继续执行下一个handler，在本例中是mainHandler。
    }
    func after(ctx iris.Context) {
        println("After the mainHandler")
    }
    func mainHandler(ctx iris.Context) {
        println("Inside mainHandler")
        // take the info from the "before" handler.
        info := ctx.Values().GetString("info")
        // write something to the client as a response.
        ctx.HTML("<h1>Response</h1>")
        ctx.HTML("<br/> Info: " + info)
        ctx.Next() // 继续下一个handler 这里是after
    }


```

## 运行

```bash
    $ go run main.go # and navigate to the http://localhost:8080
    Now listening on: http://localhost:8080
    Application started. Press CTRL+C to shut down.
    Before the mainHandler: /
    Inside mainHandler
    After the mainHandler
```

## 全局使用中间件

```go
    package main
    import "github.com/kataras/iris"
    func main() {
        app := iris.New()
        //将“before”处理程序注册为将要执行的第一个处理程序
        //在所有域的路由上。
        //或使用`UseGlobal`注册一个将跨子域触发的中间件。
        app.Use(before)

        //将“after”处理程序注册为将要执行的最后一个处理程序
        //在所有域的路由'处理程序之后。
        app.Done(after)

        // register our routes.
        app.Get("/", indexHandler)
        app.Get("/contact", contactHandler)

        app.Run(iris.Addr(":8080"))
    }
    func before(ctx iris.Context) {
         // [...]
    }
    func after(ctx iris.Context) {
        // [...]
    }
    func indexHandler(ctx iris.Context) {
        ctx.HTML("<h1>Index</h1>")
        ctx.Next() // 执行通过`Done`注册的“after”处理程序。
    }
    func contactHandler(ctx iris.Context) {
        // write something to the client as a response.
        ctx.HTML("<h1>Contact</h1>")
        ctx.Next() // 执行通过`Done`注册的“after”处理程序。
    }

```
## 中间件demo
|Middleware名称|例子地址|
|--|--|
|basic authentication|https://github.com/kataras/iris/tree/master/_examples/authentication |
|Google reCAPTCHA|https://github.com/kataras/iris/tree/master/_examples/miscellaneous/recaptcha |
|request logger|https://github.com/kataras/iris/tree/master/_examples/http_request/request-logger |
|article_id	profiling|https://github.com/kataras/iris/tree/master/_examples/miscellaneous/pprof |
|article_id	recovery|https://github.com/kataras/iris/tree/master/_examples/miscellaneous/recover |

## 中间件参考

|Middleware名称|描述|例子地址|
|--|--|--|
|jwt|中间件在传入请求的Authorization标头上检查JWT并对其进行解码|https://github.com/iris-contrib/middleware/jwt/_example |
|cors|HTTP访问控制|https://github.com/iris-contrib/middleware/cors/_example |
|secure|实现一些快速安全性的中间件获胜|https://github.com/iris-contrib/middleware/secure/_example|
|tollbooth|用于限制HTTP请求的通用中间件|https://github.com/iris-contrib/middleware/tollbooth/_examples/limit-handler |
|cloudwatch	AWS|cloudwatch指标中间件|https://github.com/iris-contrib/middleware/cloudwatch/_example |
|new relic|官方New Relic Go Agent|https://github.com/iris-contrib/middleware/newrelic/_example |
|prometheus|轻松为prometheus检测工具创建指标端点|https://github.com/iris-contrib/middleware/prometheus/_example |
|casbin|一个授权库，支持ACL，RBAC，ABAC等访问控制模型|https://github.com/iris-contrib/middleware/casbin/_examples |

