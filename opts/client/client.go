package client

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/tidwall/gjson"
	"github.com/zengzhengrong/request/config"
	"github.com/zengzhengrong/request/request"
	"github.com/zengzhengrong/request/response"
)

type DebugOption bool
type TimeoutOption time.Duration
type TransportOption struct{ http.RoundTripper }
type CheckRedirectOption func(req *http.Request, via []*http.Request) error
type TLSClientConfigOption struct{ tls.Config }
type Default struct{}
type RespBodySizeOption int64

type ClientOptions struct {
	Debug           bool
	Timeout         time.Duration
	Transport       http.RoundTripper
	CheckRedirect   func(req *http.Request, via []*http.Request) error
	TLSClientConfig TLSClientConfigOption
	RespBodySize    int64
}

type ClientOption interface {
	apply(*ClientOptions)
}

func (t TimeoutOption) apply(opts *ClientOptions) {
	opts.Timeout = time.Duration(t)
}

func (d DebugOption) apply(opts *ClientOptions) {
	opts.Debug = bool(d)
}

func (t TransportOption) apply(opts *ClientOptions) {
	opts.Transport = TransportOption{t}
}

func (c CheckRedirectOption) apply(opts *ClientOptions) {
	opts.CheckRedirect = CheckRedirectOption(c)
}

func (t TLSClientConfigOption) apply(opts *ClientOptions) {
	opts.TLSClientConfig = TLSClientConfigOption(t)
	opts.Transport = &http.Transport{TLSClientConfig: &opts.TLSClientConfig.Config}
}

func (s RespBodySizeOption) apply(opts *ClientOptions) {
	opts.RespBodySize = int64(s)
}

func (d Default) apply(opts *ClientOptions) {
	// no processing
}

// WithTransport is coustom clien transport option
func WithTransport(t http.RoundTripper) ClientOption {
	return TransportOption{t}
}

// WithInsecureSkipVerify is will override Transport
func WithInsecureSkipVerify() ClientOption {
	return &TLSClientConfigOption{tls.Config{InsecureSkipVerify: true}}
}

// WithDebug is whether or not debug
func WithDebug(debug ...bool) ClientOption {
	d := true
	if len(debug) > 0 {
		d = debug[0]
	}
	return DebugOption(d)
}

// WithTimeOut is timeout of during request
func WithTimeOut(timeout time.Duration) ClientOption {
	return TimeoutOption(timeout)
}

// If CheckRedirect is nil, the Client uses its default policy,
// which is to stop after 10 consecutive requests.
func WithCheckRedirect(f func(req *http.Request, via []*http.Request) error) ClientOption {
	return CheckRedirectOption(f)
}

// WithDefault is use defaut client
func WithDefault() ClientOption {
	return Default{}
}

// RespBodySize
func WithPreRespBodySize(size ...int64) RespBodySizeOption {
	var s int64
	if len(size) > 0 {
		s = size[0]
	}
	return RespBodySizeOption(s)
}

type Client struct {
	Opts       *ClientOptions
	HttpClient *http.Client
}

func NewClient(opts ...ClientOption) *Client {
	options := &ClientOptions{
		Debug:         config.SetDefaultDebug(),
		Timeout:       config.DefaultTimeout,
		Transport:     http.DefaultTransport,
		CheckRedirect: config.DefaultCheckRedirect,
	}
	for _, o := range opts {
		o.apply(options)
	}

	client := &http.Client{
		Transport:     options.Transport,
		CheckRedirect: options.CheckRedirect, // 获取301重定向
		Timeout:       options.Timeout,
	}
	return &Client{
		Opts:       options,
		HttpClient: client,
	}
}

// Do is ShortCut http client do method
func (client *Client) Do(r *request.Request) (*http.Response, error) {

	if client.Opts.Debug {
		// DEBUG mode request >> connect >> client(option) >> response(option)
		spew.Dump(r.Opts)                   // print request options
		clientTrace := defaultclientTrace() // use default trace if open debug
		req := r.HttpReq.WithContext(httptrace.WithClientTrace(r.HttpReq.Context(), clientTrace))
		now := time.Now()
		resp, err := client.HttpClient.Do(req)
		e := time.Since(now)
		elapsed := struct{ elapsed time.Duration }{elapsed: e}
		spew.Dump(elapsed) // print cost time
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()                               //  must close
		resp.Body = io.NopCloser(bytes.NewBuffer(body)) // rewrite resp Body
		spew.Dump(gjson.ParseBytes(body))               // print response
		if os.Getenv("REQUEST_CLIENT_DEBUG") != "" {
			spew.Dump(client.Opts) // print client options
		}
		return resp, err
	}
	resp, err := client.HttpClient.Do(r.HttpReq)
	return resp, err

}

func (client *Client) Req(method string, url string, postbody any, args ...map[string]string) response.Response {
	var body []byte
	query, header := request.Getqueryheader(args...)
	r, err := request.NewReuqest(
		method,
		url,
		request.WithBody(postbody),
		request.WithQuery(query),
		request.WithHeader(header),
	)
	if err != nil {
		return response.Response{Resp: nil, Body: nil, Err: err}
	}
	resp, err := client.Do(r)
	if err != nil {
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	defer resp.Body.Close()

	// use ReadFull/Copy Instead of ReadAll reduce memory allocation
	if client.Opts.RespBodySize != 0 {
		if client.Opts.RespBodySize < resp.ContentLength {
			// if body size less than content-length
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return response.Response{Resp: resp, Body: nil, Err: err}
			}
			return response.Response{Resp: resp, Body: body, Err: nil}
		}
		buf := make([]byte, 0, client.Opts.RespBodySize)
		buffer := bytes.NewBuffer(buf)
		_, err := io.Copy(buffer, resp.Body)
		if err != nil {
			return response.Response{Resp: resp, Body: nil, Err: err}
		}
		temp := buffer.Bytes()
		length := len(temp)
		if cap(temp) > (length + length/10) {
			body = make([]byte, length)
			copy(body, temp)
		} else {
			body = temp
		}

		return response.Response{Resp: resp, Body: body, Err: nil}
	} else {
		body = make([]byte, resp.ContentLength)
	}
	_, err = io.ReadFull(resp.Body, body)
	if err != nil {
		if err == io.EOF {
			return response.Response{Resp: resp, Body: body, Err: nil}
		}
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	return response.Response{Resp: resp, Body: body, Err: nil}
}

// ReqRaw just warp the http.Response do not read body , must Close the body after you read the body
func (client *Client) ReqRaw(method string, url string, postbody any, args ...map[string]string) response.Response {
	query, header := request.Getqueryheader(args...)
	r, err := request.NewReuqest(
		method,
		url,
		request.WithBody(postbody),
		request.WithQuery(query),
		request.WithHeader(header),
	)
	if err != nil {
		return response.Response{Resp: nil, Body: nil, Err: err}
	}
	resp, err := client.Do(r)
	if err != nil {
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	return response.Response{Resp: resp, Body: nil, Err: nil}

}

// GET is reuse client
func (client *Client) GET(url string, args ...map[string]string) response.Response {
	return client.Req(http.MethodGet, url, nil, args...)
}

// POST is shortcut post method with json and client
func (client *Client) POST(url string, postbody any, args ...map[string]string) response.Response {
	return client.Req(http.MethodPost, url, postbody, args...)

}

// PUT is shortcut post method with json and client
func (client *Client) PUT(url string, postbody any, args ...map[string]string) response.Response {
	return client.Req(http.MethodPut, url, postbody, args...)

}

// PATCH is shortcut post method with json and client
func (client *Client) PATCH(url string, postbody any, args ...map[string]string) response.Response {
	return client.Req(http.MethodPatch, url, postbody, args...)

}

// DELETE is shortcut post method with json and client
func (client *Client) DELETE(url string, postbody any, args ...map[string]string) response.Response {
	return client.Req(http.MethodDelete, url, postbody, args...)

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
