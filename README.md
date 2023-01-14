# request
wrap net/http client  use options mode


封装 net/http 客户端 从zzgo 仓库独立出来，使用选项式编程


## 用法

可以查看 [测试用例](https://github.com/zengzhengrong/request/blob/main/test/http_test.go)



## 快捷请求
[便捷函数](https://github.com/zengzhengrong/request/blob/main/curl/curl.go)

支持 GET POST PUT PATCH DELETE 复用会话 上传文件


### 直接绑定结构体
GET
```
	result := &Result{}
	err := curl.GETBind(result, "https://httpbin.org/get", testquery(), testheader())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
```
POST json
```
	result := &Result{}
	err := curl.POSTBind(result, "https://httpbin.org/post", testjsonbody(), testquery(), testheader())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
```

POST form

```
	result := &Result{}
	err := curl.POSTFormBind(result, "https://httpbin.org/post", testformbody(), testquery(), testheader())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
```


## 复用会话

```
	client := client.NewClient(
		client.WithDebug(true),
		client.WithTimeOut(10*time.Second),
	)
	resp1 := client.GET("https://httpbin.org/get")
	resp2 := client.POST("https://httpbin.org/post", testjsonbody(), testquery(), testheader())
	fmt.Println(resp1.GetBodyString())
	fmt.Println(resp2.GetBodyString())
```

开启debug 模式可以看到打印  Reused: (bool) true 字样 说明 第二个POST请求复用会话成功


## 流水线请求

pipline
```
	c := client.NewClient(client.WithDefault())
	p := pipline.NewPipLine(
		pipline.WithParall(true),
		pipline.WithClient(c),
		pipline.WithIn(func(ctx context.Context, cli *client.Client) ([]byte, error) {
			resp := curl.ClientGET(cli, "https://httpbin.org/get", testquery(), testheader())
			if resp.GetError() != nil {
				return nil, resp.GetError()
			}
			return resp.Body, nil
		}, func(ctx context.Context, cli *client.Client) ([]byte, error) {
			resp := curl.ClientPOST(cli, "https://httpbin.org/post", testjsonbody(), testquery(), testheader())
			if resp.GetError() != nil {
				return nil, resp.GetError()
			}
			return resp.Body, nil
		}),
		pipline.WithOut(func(ctx context.Context, cli *client.Client, Ins ...[]byte) request.Response {
			r1 := gjson.GetBytes(Ins[0], "args.a").String()
			r2 := gjson.GetBytes(Ins[1], "json").Value()
			body := struct {
				R1 string
				R2 any
			}{
				R1: r1,
				R2: r2,
			}
			b, _ := json.Marshal(body)
			resp := curl.ClientPOST(cli, "https://httpbin.org/post", b, testquery(), testheader())
			return resp
		}),
	)
	resp := p.Result()
	fmt.Println(string(resp.Body))
```

pipline.WithParall(true) 并发请求pipline.WithIn 的函数  
pipline.WithClient(c) 流水线的所有请求复用会话  
pipline.WithIn 获取请求当作pipline.WithOut的入参  
pipline.WithOut 跟进WithIn 组合请求获取最终结果  
p.Result() 运行整个流水线获取pipline.WithOut响应  


## 命令行工具

zurl 主要解决在kubernetes部署接口应用的时候用来做 上游依赖检查(init container) 目前网上通常做法例如

```
...
      initContainers:
      - name: wait-canal-admin
        image: zengzhengrong889/box:1.0
        command: 
          - sh
          - -c
          - until curl -s -o /dev/null canal-admin.canal; do echo waiting for canal-admin; sleep 2; done; echo done
```
循环调用curl 直到上游服务canal-admin.canal有返回 才运行主容器，但是往往 canal-admin.canal 这个service name 能访问，却实际 这个服务并没有完全启动起来，
通常 直接就输出 了done，然后就开始运行主容器，主容器 启动后检查到上游服务并不能有效访问，就会退出，导致pod 会进行重启数遍 直到 上游应用有效访问，而zurl就是解决这个问题


将上面yaml 替换如下 即可避免主容器重启

```
      initContainers:
      - name: wait-canal-admin
        image: zengzhengrong889/zurl:latest
        args:
          - "--url"
          - "http://canal-admin.canal/api/v1/login"
          - "--retry"
          - "20"
          - "--debug"
          - "true"
```
--url 指定检查上游服务是否可用的接口  
--retry  重试次数超过就init container就会失败  
--debug 开启 debug模式  

debug 模式会打印请求信息以及dns信息
```
(*request.ReqOptions)(0xc000202800)({
Method: (string) (len=3) "GET",
Url: (string) (len=38) "http://canal-admin.canal/api/v1/login?",
ContentType: (string) (len=16) "application/json",
Header: (map[string]string) (len=1) {
(string) (len=12) "Content-Type": (string) (len=16) "application/json"},
Body: (io.Reader) <nil>,
RawBody: (interface {}) <nil>,
Query: (string) "",
Context: (context.Context) <nil>})
(string) (len=20) "canal-admin.canal:80"
(httptrace.DNSStartInfo) {
Host: (string) (len=17) "canal-admin.canal"}(httptrace.DNSDoneInfo) {
Addrs: ([]net.IPAddr) (len=1 cap=1) {(net.IPAddr) 10.101.155.207},
Err: (error) <nil>,
Coalesced: (bool) false}
(struct { Reused bool }) {Reused: (bool) false}
(struct { elapsed time.Duration }) {elapsed: (time.Duration) 435.931ms}
(gjson.Result) {"code":50014,"message":"Expired token","data":null}
Math StatusCode Success 200,200
Success match condition , exit ...
```

目前只支持get 方法，可以自定义添加头```--add-header```和查询参数```--add-query```

命令行示例
```
# 默认检查状态码是否200
zurl --url https://httpbin.org/get --retry 2 
 # 指定检查状态码和设置retry间隔时间，以及header 头的内容
zurl --url https://httpbin.org/get --retry 2 --expect-statuscode 200 --interval 2 --expect-header Content-Type=application/json
# 指定检查状态码和设置retry间隔时间，以及json 的内容(value字符串匹配)，json key的路径可以用 xx.xx指定
zurl --url https://httpbin.org/get --retry 2 --expect-statuscode 200 --interval 2 --expect-json url=https://httpbin.org/get 
```



