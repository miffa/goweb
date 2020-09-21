# iris restful 路由参数解析方式说明

```go
 app := iris.New()
// 您可以使用“string”类型，该类型对于可以是任何内容的单个路径参数有效
app.Get("/username/{name}", func(ctx iris.Context) {
    ctx.Writef("Hello %s", ctx.Params().Get("name"))
}) //  {name:string}

//注册我们的第一个附加到int macro类型的宏
// "min" = 当前函数名字
// "minValue" = 函数的参数
// func(string) bool = macro's 的路径参数评估器，这在服务时执行
// 用户使用min（...）macros参数函数请求包含：int macros类型的路径。
app.Macros().Int.RegisterFunc("min", func(minValue int) func(string) bool {
    // 在此之前做任何事情[...]
    //在这种情况下，我们不需要做任何事情
    return func(paramValue string) bool {
        n, err := strconv.Atoi(paramValue)
        if err != nil {
            return false
        }
        return n >= minValue
    }
})
// http://localhost:8080/profile/id>=1
// 这将抛出404，即使它被发现为路线 : /profile/0, /profile/blabla, /profile/-1
// macros 参数函数当然是可选的

app.Get("/profile/{id:int min(1)}", func(ctx iris.Context) {
    //  第二个参数是错误的，因为我们使用 macros 它总是为nil
    // 验证已经发生了.
    id, _ := ctx.Params().GetInt("id")
    ctx.Writef("Hello id: %d", id)
})

// 更改每个路由的macros评估程序的错误代码：
app.Get("/profile/{id:int min(1)}/friends/{friendid:int min(1) else 504}", func(ctx iris.Context) {
    id, _ := ctx.Params().GetInt("id")
    friendid, _ := ctx.Params().GetInt("friendid")
    ctx.Writef("Hello id: %d looking for friend id: ", id, friendid)
}) // 如果没有传递所有路由的macros，这将抛出504错误代码而不是404.

// http://localhost:8080/game/a-zA-Z/level/0-9
//记住，字母只是小写或大写字母。
app.Get("/game/{name:alphabetical}/level/{level:int}", func(ctx iris.Context) {
    ctx.Writef("name: %s | level: %s", ctx.Params().Get("name"), ctx.Params().Get("level"))
})

//让我们使用一个简单的自定义regexp来验证单个路径参数
//它的值只是小写字母。

// http://localhost:8080/lowercase/anylowercase
app.Get("/lowercase/{name:string regexp(^[a-z]+)}", func(ctx iris.Context) {
    ctx.Writef("name should be only lowercase, otherwise this handler will never executed: %s", ctx.Params().Get("name"))
})

// http://localhost:8080/single_file/app.js
app.Get("/single_file/{myfile:file}", func(ctx iris.Context) {
    ctx.Writef("file type validates if the parameter value has a form of a file name, got: %s", ctx.Params().Get("myfile"))
})

// http://localhost:8080/myfiles/any/directory/here/
// 这是唯一接受任意数量路径段的macro类型。
app.Get("/myfiles/{directory:path}", func(ctx iris.Context) {
    ctx.Writef("path type accepts any number of path segments, path after /myfiles/ is: %s", ctx.Params().Get("directory"))
})

app.Run(iris.Addr(":8080"))
}


```
