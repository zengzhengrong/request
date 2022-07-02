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

