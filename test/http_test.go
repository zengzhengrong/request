package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/zengzhengrong/request"
	"github.com/zengzhengrong/request/curl"
	"github.com/zengzhengrong/request/opts/client"
	"github.com/zengzhengrong/request/opts/pipline"
)

var (
	query              map[string]string
	header             map[string]string
	jsonbody           []byte
	formbody           map[string]string
	defaultclienttrace *httptrace.ClientTrace
)

type Result struct {
	Args    Args    `json:"args"`
	Headers Headers `json:"headers"`
	Origin  string  `json:"origin"`
	URL     string  `json:"url"`
	Form    Form    `json:"form"`
}
type Args struct {
	A string `json:"a"`
	B string `json:"b"`
}
type Form struct {
	AA string `json:"aa"`
	BA string `json:"ba"`
}
type Headers struct {
	A              string `json:"A"`
	AcceptEncoding string `json:"Accept-Encoding"`
	B              string `json:"B"`
	Host           string `json:"Host"`
	UserAgent      string `json:"User-Agent"`
	XAmznTraceID   string `json:"X-Amzn-Trace-Id"`
}

func testheader() map[string]string {
	return map[string]string{
		"A": "a",
		"B": "b",
	}
}

func testquery() map[string]string {
	return map[string]string{
		"a": "1",
		"b": "2",
	}
}

func testjsonbody() []byte {
	body := map[string]string{
		"aa": "1",
		"ba": "2",
	}
	b, _ := json.Marshal(body)
	return b
}

func testformbody() map[string]string {
	return map[string]string{
		"aa": "1",
		"ba": "2",
	}
}
func defaultclientTrace() (clientTrace *httptrace.ClientTrace) {

	clientTrace = &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			spew.Dump(info)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			spew.Dump(info)
		},
		GetConn: func(hostPort string) {
			spew.Dump(hostPort)
		},
		GotConn: func(gci httptrace.GotConnInfo) {
			if os.Getenv("REQUEST_CONN_DEBUG") != "" {
				spew.Dump(gci)
			} else {
				reused := struct {
					Reused bool
				}{gci.Reused}
				spew.Dump(reused)
			}

		},
	}

	return clientTrace
}

func TestMain(m *testing.M) {
	fmt.Println("初始化参数")
	query = testquery()
	header = testheader()
	jsonbody = testjsonbody()
	formbody = testformbody()
	defaultclienttrace = defaultclientTrace()
	m.Run()
}

func TestHtppQuery(t *testing.T) {
	url := "https://httpbin.org?"
	args := map[string]string{
		"a": "1",
		"b": "2",
	}
	result := request.HttpBuildQuery(args)
	fmt.Println(result)
	url = url + result
	fmt.Println(url)
	fmt.Println(strings.Index(url, "还"))
}

func TestRequest(t *testing.T) {
	h := testheader()
	q := testquery()
	body := testjsonbody()
	r, err := request.NewReuqest(
		http.MethodGet,
		"https://httpbin.org/get",
		request.WithHeader(h),
		request.WithBody(body),
		request.WithQuery(q),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
}

func TestClient(t *testing.T) {
	h := testheader()
	q := testquery()
	body := testjsonbody()
	r, err := request.NewReuqest(
		http.MethodPost,
		"https://httpbin.org/post",
		request.WithHeader(h),
		request.WithBody(body),
		request.WithQuery(q),
	)

	if err != nil {
		panic(err)
	}
	client := client.NewClient(
		client.WithDebug(true),
		client.WithTimeOut(10*time.Second),
	)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	res, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	if err := resp.Body.Close(); err != nil {
		panic(err)
	}

	fmt.Println(string(res))
	fmt.Println(resp.Close)

	r2 := r.Clone()
	fmt.Println(r2)
	resp, err = client.Do(r2)
	if err != nil {
		panic(err)
	}
	res, err = io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(res))
	fmt.Println(resp.Close)
}

func TestGET(t *testing.T) {
	h := testheader()
	q := testquery()
	body := testjsonbody()
	r, err := request.NewReuqest(
		http.MethodGet,
		"https://httpbin.org/get",
		request.WithHeader(h),
		request.WithBody(body),
		request.WithQuery(q),
	)
	if err != nil {
		panic(err)
	}
	client := client.NewClient(
		client.WithTimeOut(10 * time.Second),
	)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	resbyte, _ := io.ReadAll(resp.Body)

	fmt.Println(string(resbyte))
	assert.Equal(t, "200 OK", resp.Status)

	resp.Body.Close()
}

func TestPOST(t *testing.T) {
	h := testheader()
	q := testquery()
	body := testjsonbody()
	r, err := request.NewReuqest(
		http.MethodPost,
		"http://127.0.0.1:8081/post",
		request.WithHeader(h),
		request.WithBody(body),
		request.WithQuery(q),
	)
	if err != nil {
		panic(err)
	}
	client := client.NewClient(
		client.WithDebug(),
		client.WithTimeOut(10*time.Second),
	)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	tobody := make([]byte, resp.ContentLength)

	n, err := io.ReadFull(resp.Body, tobody)
	if err != nil {
		panic(err)
	}
	s := string(tobody)
	fmt.Println(n)
	fmt.Println(s)
	fmt.Println(len(s))
	assert.Equal(t, "200 OK", resp.Status)

	err = resp.Body.Close()
	if err != nil {
		panic(err)
	}
	err = resp.Body.Close()
	if err != nil {
		panic(err)
	}
}

func TestPUT(t *testing.T) {
	h := testheader()
	q := testquery()
	body := testjsonbody()
	r, err := request.NewReuqest(
		http.MethodPut,
		"https://httpbin.org/put",
		request.WithHeader(h),
		request.WithBody(body),
		request.WithQuery(q),
	)
	if err != nil {
		panic(err)
	}
	client := client.NewClient(
		client.WithDebug(true),
		client.WithTimeOut(10*time.Second),
	)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	resbyte, _ := io.ReadAll(resp.Body)

	fmt.Println(string(resbyte))
	assert.Equal(t, "200 OK", resp.Status)

	resp.Body.Close()
}

func TestPATCH(t *testing.T) {
	h := testheader()
	q := testquery()
	body := testjsonbody()
	r, err := request.NewReuqest(
		http.MethodPatch,
		"https://httpbin.org/patch",
		request.WithHeader(h),
		request.WithBody(body),
		request.WithQuery(q),
	)
	if err != nil {
		panic(err)
	}
	client := client.NewClient(
		client.WithDebug(true),
		client.WithTimeOut(10*time.Second),
	)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	resbyte, _ := io.ReadAll(resp.Body)

	fmt.Println(string(resbyte))
	assert.Equal(t, "200 OK", resp.Status)

	resp.Body.Close()
}

func TestDELETE(t *testing.T) {
	h := testheader()
	q := testquery()
	body := testjsonbody()
	r, err := request.NewReuqest(
		http.MethodDelete,
		"https://httpbin.org/delete",
		request.WithHeader(h),
		request.WithBody(body),
		request.WithQuery(q),
	)
	if err != nil {
		panic(err)
	}
	client := client.NewClient(
		client.WithDebug(true),
		client.WithTimeOut(10*time.Second),
	)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	resbyte, _ := io.ReadAll(resp.Body)

	fmt.Println(string(resbyte))
	assert.Equal(t, "200 OK", resp.Status)

	resp.Body.Close()
}

func TestShortCutGET(t *testing.T) {
	res := curl.GET("https://httpbin.org/get", testquery(), testheader())
	fmt.Println(string(res.Body))
	fmt.Println(res.OK())
	fmt.Println(res.OKByJsonKey("args", 1))
	result := &Headers{}
	res.GetKeyStruct(result, "headers")
	fmt.Println(result)
}

func TestGETBind(t *testing.T) {
	result := &Result{}
	err := curl.GETBind(result, "https://httpbin.org/get", testquery(), testheader())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func TestShortCutPOST(t *testing.T) {

	res := curl.POST("https://httpbin.org/post", testjsonbody(), testquery(), testheader())
	fmt.Println(res.OK())
	fmt.Println(res.GetBodyString())
}

func TestShortCutPOSTForm(t *testing.T) {

	res := curl.POSTForm("https://httpbin.org/post", testformbody(), testquery(), testheader())
	fmt.Println(res.OK())
	fmt.Println(res.GetBodyString())
}

func TestShortCutPOSTBind(t *testing.T) {
	result := &Result{}
	err := curl.POSTBind(result, "https://httpbin.org/post", testjsonbody(), testquery(), testheader())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func TestShortCutPOSTFormBind(t *testing.T) {
	result := &Result{}
	err := curl.POSTFormBind(result, "https://httpbin.org/post", testformbody(), testquery(), testheader())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func TestNewPipLine(t *testing.T) {
	err := os.Setenv("REQUEST_DEBUG", "1")
	if err != nil {
		panic(err)
	}
	c := client.NewClient(client.WithDefault())
	p := pipline.NewPipLine(
		pipline.WithParall(true),
		pipline.WithClient(c),
		pipline.WithIn(func(ctx context.Context, cli *client.Client) ([]byte, error) {
			resp := cli.GET("https://httpbin.org/get", query, header)
			if resp.GetError() != nil {
				return nil, resp.GetError()
			}
			return resp.Body, nil
		}, func(ctx context.Context, cli *client.Client) ([]byte, error) {
			resp := cli.POST("https://httpbin.org/post", jsonbody, query, header)
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
			resp := cli.POST("https://httpbin.org/post", b, query, header)
			return resp
		}),
	)
	resp := p.Result()
	if resp.Err != nil {
		panic(resp.Err)
	}
	fmt.Println(string(resp.Body))
}

func TestDEBUG(t *testing.T) {
	err := os.Setenv("REQUEST_DEBUG", "1")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("REQUEST_RESPONSE_DEBUG", "1")
	if err != nil {
		panic(err)
	}
	result := &Result{}
	err = curl.GETBind(result, "https://httpbin.org/get", testquery(), testheader())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func TestWithContext(t *testing.T) {
	h := testheader()
	q := testquery()
	body := testjsonbody()
	clientTrace := defaultclientTrace()
	ctx := httptrace.WithClientTrace(context.Background(), clientTrace)
	r, err := request.NewReuqest(
		http.MethodGet,
		"https://httpbin.org/get",
		request.WithHeader(h),
		request.WithBody(body),
		request.WithQuery(q),
		request.WithContext(ctx),
	)
	if err != nil {
		panic(err)
	}
	client := client.NewClient(
		client.WithDebug(false),
		client.WithTimeOut(10*time.Second),
	)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	resbyte, _ := io.ReadAll(resp.Body)

	fmt.Println(string(resbyte))
	assert.Equal(t, "200 OK", resp.Status)

	resp.Body.Close()
}

func TestReuseClient(t *testing.T) {
	client := client.NewClient(
		client.WithPreRespBodySize(500),
		client.WithTimeOut(10*time.Second),
	)
	resp1 := client.GET("https://httpbin.org/get", query, header)
	// resp2 := client.POST("https://httpbin.org/post", jsonbody, query, header)
	// resp3 := client.PATCH("https://httpbin.org/patch", jsonbody, query, header)
	// resp4 := client.PUT("https://httpbin.org/put", jsonbody, query, header)
	// resp5 := client.DELETE("https://httpbin.org/delete", query, header)
	fmt.Println(resp1.GetBodyString())
	// fmt.Println(resp2.GetBodyString())
	// fmt.Println(resp3.GetBodyString())
	// fmt.Println(resp4.GetBodyString())
	// fmt.Println(resp5.GetBodyString())

}
