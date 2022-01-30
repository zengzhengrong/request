package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/zengzhengrong/request"
	"github.com/zengzhengrong/request/curl"
	"github.com/zengzhengrong/request/opts/client"
	"github.com/zengzhengrong/request/opts/pipline"
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

	res, err := ioutil.ReadAll(resp.Body)

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
	res, err = ioutil.ReadAll(resp.Body)
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

func TestPOST(t *testing.T) {
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
	resbyte, _ := io.ReadAll(resp.Body)

	fmt.Println(string(resbyte))
	assert.Equal(t, "200 OK", resp.Status)

	resp.Body.Close()
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
	fmt.Println(resp.OK())
	fmt.Println(resp.GetError())
}
