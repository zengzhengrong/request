package request

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/spf13/cast"
)

type Request struct {
	Opts    *ReqOptions   `json:"opts"`
	HttpReq *http.Request `json:"http_req"`
}

// http request options
type ReqOptions struct {
	Method      string
	Url         string
	ContentType string
	Header      map[string]string
	Body        io.Reader
	RawBody     any
	Query       string
}
type ReqOption interface {
	apply(*ReqOptions)
}

type ContentTypeOption string
type HeaderOption map[string]string
type BodyOption struct {
	io.Reader
	raw any
}
type QueryOption string

func (c ContentTypeOption) apply(opts *ReqOptions) {
	opts.ContentType = string(c)
}

func (h HeaderOption) apply(opts *ReqOptions) {
	opts.Header = h
}
func (q QueryOption) apply(opts *ReqOptions) {
	opts.Query = string(q)
	if strings.Index(opts.Url, "?") != -1 {
		opts.Url = opts.Url + "&" + string(q)
	} else {
		opts.Url = opts.Url + "?" + string(q)
	}

}

func (b BodyOption) apply(opts *ReqOptions) {
	opts.Body = b.Reader
	opts.RawBody = b.raw
}

// WithContentType is set http client content-type
func WithContentType(c string) ReqOption {
	return ContentTypeOption(c)
}

func WithHeader(h map[string]string) ReqOption {
	return HeaderOption(h)
}

func WithBody(body any) ReqOption {
	var reqBody io.Reader
	switch v := body.(type) {
	case io.Reader:
		reqBody = v
	case string:
		reqBody = strings.NewReader(v)
	case []byte:
		reqBody = strings.NewReader(string(v))
	case map[string]string:
		val := url.Values{}
		for k, v := range v {
			val.Set(k, v)
		}
		reqBody = strings.NewReader(val.Encode())
	case map[string]any:
		val := url.Values{}
		for k, v := range v {
			val.Set(k, v.(string))
		}
		reqBody = strings.NewReader(val.Encode())
	}
	return BodyOption{reqBody, body}
}

func WithQuery(query map[string]string, sortAsc ...bool) ReqOption {
	return QueryOption(HttpBuildQuery(query, sortAsc...))
}

func (r *Request) String() string {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.Encode(r.Opts)
	return bf.String()
}

func (r *Request) Clone() *Request {
	b := WithBody(r.Opts.RawBody)
	b.apply(r.Opts)
	req, _ := http.NewRequest(r.Opts.Method, r.Opts.Url, r.Opts.Body)
	for k, v := range r.Opts.Header {
		req.Header.Set(k, v)
	}
	newRequest := &Request{
		Opts:    r.Opts,
		HttpReq: req,
	}
	return newRequest
}

func NewReuqest(method string, url string, opts ...ReqOption) (*Request, error) {
	options := &ReqOptions{
		Method:      method,
		Url:         url,
		ContentType: DefaultContentType,
	}
	for _, o := range opts {
		o.apply(options)
	}
	if options.ContentType != "" {
		options.Header["Content-Type"] = options.ContentType
	}
	r, err := http.NewRequest(options.Method, options.Url, options.Body)
	if err != nil {
		return nil, err
	}
	for k, v := range options.Header {
		r.Header.Set(k, v)
	}

	return &Request{
		Opts:    options,
		HttpReq: r,
	}, nil
}

// 组建query请求参数,sortAsc true为小到大,false为大到小,nil不排序  a=123&b=321
func HttpBuildQuery(args map[string]string, sortAsc ...bool) string {
	str := ""
	if len(args) == 0 {
		return str
	}
	if len(sortAsc) > 0 {
		keys := make([]string, 0, len(args))
		for k := range args {
			keys = append(keys, k)
		}
		if sortAsc[0] {
			sort.Strings(keys)
		} else {
			sort.Sort(sort.Reverse(sort.StringSlice(keys)))
		}
		for _, k := range keys {
			str += "&" + k + "=" + args[k]
		}
	} else {
		for k, v := range args {
			str += "&" + k + "=" + v
		}
	}
	return str[1:]
}

func HttpBuild(body interface{}, sortAsc ...bool) string {
	params := map[string]string{}
	if args, ok := body.(map[string]interface{}); ok {
		for k, v := range args {
			params[k] = cast.ToString(v)
		}
		return HttpBuildQuery(params, sortAsc...)
	}
	if args, ok := body.(map[string]string); ok {
		for k, v := range args {
			params[k] = cast.ToString(v)
		}
		return HttpBuildQuery(params, sortAsc...)
	}
	if args, ok := body.(map[string]int); ok {
		for k, v := range args {
			params[k] = cast.ToString(v)
		}
		return HttpBuildQuery(params, sortAsc...)
	}
	return cast.ToString(body)
}

// GetFileOrFiles is 获取单个文件或者多个文件
// 多个文件以file_{0}开头读取form-data，最多20个 在map中 键名以文件名为键名
// 单文件以file读取form-data
func GetFileOrFiles(req *http.Request) (bool, map[string]*multipart.FileHeader, error) {
	var name string
	var IsSingle bool
	files := make(map[string]*multipart.FileHeader, 20)
	for i := 0; i < MaxUploadThreads; i++ {
		name = "file_" + cast.ToString(i)
		fs, err := fromfile(req, name)
		if err != nil && i == 0 {
			IsSingle = true
			break

		}
		if err != nil {
			// Not found break
			break
		}
		files[fs.Filename] = fs
	}
	if IsSingle && len(files) == 0 {
		sf, err := fromfile(req, "file")
		if err != nil {
			return IsSingle, files, err
		}
		files[sf.Filename] = sf
		return IsSingle, files, nil
	}
	return IsSingle, files, nil
}

func fromfile(req *http.Request, name string) (*multipart.FileHeader, error) {
	if req.MultipartForm == nil {
		if err := req.ParseMultipartForm(DefaultMultipartMemory); err != nil {
			return nil, err
		}
	}
	f, fh, err := req.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, err
}
