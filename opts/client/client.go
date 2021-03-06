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
	"github.com/zengzhengrong/request"
)

type DebugOption bool
type TimeoutOption time.Duration
type TransportOption struct{ http.RoundTripper }
type CheckRedirectOption func(req *http.Request, via []*http.Request) error
type TLSClientConfigOption struct{ tls.Config }
type Default struct{}

type ClientOptions struct {
	Debug           bool
	Timeout         time.Duration
	Transport       http.RoundTripper
	CheckRedirect   func(req *http.Request, via []*http.Request) error
	TLSClientConfig TLSClientConfigOption
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

func (d Default) apply(opts *ClientOptions) {
	// no processing
}

func WithTransport(t http.RoundTripper) ClientOption {
	return TransportOption{t}
}

// WithInsecureSkipVerify is will override Transport
func WithInsecureSkipVerify() ClientOption {
	return &TLSClientConfigOption{tls.Config{InsecureSkipVerify: true}}
}

func WithDebug(debug bool) ClientOption {
	return DebugOption(debug)
}
func WithTimeOut(timeout time.Duration) ClientOption {
	return TimeoutOption(timeout)
}

func WithCheckRedirect(f func(req *http.Request, via []*http.Request) error) ClientOption {
	return CheckRedirectOption(f)
}

func WithDefault() ClientOption {
	return Default{}
}

type Client struct {
	Opts       *ClientOptions
	HttpClient *http.Client
}

func NewClient(opts ...ClientOption) *Client {
	options := &ClientOptions{
		Debug:         request.SetDefaultDebug(),
		Timeout:       request.DefaultTimeout,
		Transport:     http.DefaultTransport,
		CheckRedirect: request.DefaultCheckRedirect,
	}
	for _, o := range opts {
		o.apply(options)
	}

	client := &http.Client{
		Transport:     options.Transport,
		CheckRedirect: options.CheckRedirect, // ??????301?????????
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
		spew.Dump(r.Opts)
		clientTrace := defaultclientTrace() // use default trace if open debug
		req := r.HttpReq.WithContext(httptrace.WithClientTrace(r.HttpReq.Context(), clientTrace))
		resp, err := client.HttpClient.Do(req)
		if os.Getenv("REQUEST_CLIENT_DEBUG") != "" {
			spew.Dump(client.Opts)
		}
		if os.Getenv("REQUEST_RESPONSE_DEBUG") != "" {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()                               //  must close
			resp.Body = io.NopCloser(bytes.NewBuffer(body)) // rewrite resp Body
			spew.Dump(gjson.ParseBytes(body))
		}
		return resp, err
	}
	resp, err := client.HttpClient.Do(r.HttpReq)
	return resp, err

}

func (client *Client) Req(method string, url string, postbody any, args ...map[string]string) request.Response {
	query, header := request.Getqueryheader(args...)
	r, err := request.NewReuqest(
		method,
		url,
		request.WithBody(postbody),
		request.WithQuery(query),
		request.WithHeader(header),
	)
	if err != nil {
		return request.Response{Resp: nil, Body: nil, Err: err}
	}
	resp, err := client.Do(r)
	if err != nil {
		return request.Response{Resp: resp, Body: nil, Err: err}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return request.Response{Resp: resp, Body: nil, Err: err}
	}
	resp.Body.Close()
	return request.Response{Resp: resp, Body: body, Err: nil}

}

// ReqRaw just warp the http.Response do not read body , must Close the body after you read the body
func (client *Client) ReqRaw(method string, url string, postbody any, args ...map[string]string) request.Response {
	query, header := request.Getqueryheader(args...)
	r, err := request.NewReuqest(
		method,
		url,
		request.WithBody(postbody),
		request.WithQuery(query),
		request.WithHeader(header),
	)
	if err != nil {
		return request.Response{Resp: nil, Body: nil, Err: err}
	}
	resp, err := client.Do(r)
	if err != nil {
		return request.Response{Resp: resp, Body: nil, Err: err}
	}
	return request.Response{Resp: resp, Body: nil, Err: nil}

}

// GET is reuse client
func (client *Client) GET(url string, args ...map[string]string) request.Response {
	return client.Req(http.MethodGet, url, nil, args...)
}

// POST is shortcut post method with json and client
func (client *Client) POST(url string, postbody any, args ...map[string]string) request.Response {
	return client.Req(http.MethodPost, url, postbody, args...)

}

// PUT is shortcut post method with json and client
func (client *Client) PUT(url string, postbody any, args ...map[string]string) request.Response {
	return client.Req(http.MethodPut, url, postbody, args...)

}

// PATCH is shortcut post method with json and client
func (client *Client) PATCH(url string, postbody any, args ...map[string]string) request.Response {
	return client.Req(http.MethodPatch, url, postbody, args...)

}

// DELETE is shortcut post method with json and client
func (client *Client) DELETE(url string, postbody any, args ...map[string]string) request.Response {
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
