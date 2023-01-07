package curl

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/zengzhengrong/request/config"
	"github.com/zengzhengrong/request/opts/client"
	"github.com/zengzhengrong/request/request"
	"github.com/zengzhengrong/request/response"
)

// GET is ShortCut get http method but not reuse tpc connect
// The first args[0] is query , args[1] is header
func GET(url string, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.Req(http.MethodGet, url, nil, args...)

}

// GETRaw is response body not close
func GETRaw(url string, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.ReqRaw(http.MethodGet, url, nil, args...)

}

// POST is shortcut post method with json
func POST(url string, postbody any, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.Req(http.MethodPost, url, postbody, args...)

}

// POSTRaw is response body not close
func POSTRaw(url string, postbody any, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.ReqRaw(http.MethodPost, url, postbody, args...)

}

// PUT is shortcut post method with json
func PUT(url string, postbody any, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.Req(http.MethodPut, url, postbody, args...)
}

// PUTRaw is response body not close
func PUTRaw(url string, postbody any, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.ReqRaw(http.MethodPut, url, postbody, args...)
}

// Patch is shortcut post method with json
func PATCH(url string, postbody any, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.Req(http.MethodPatch, url, postbody, args...)
}

// PATCHRaw is response body not close
func PATCHRaw(url string, postbody any, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.ReqRaw(http.MethodPatch, url, postbody, args...)
}

// Delete is shortcut post method with json
func DELETE(url string, postbody any, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.Req(http.MethodDelete, url, postbody, args...)
}

// DELETERaw is response body not close
func DELETERaw(url string, postbody any, args ...map[string]string) response.Response {
	client := client.NewClient(client.WithDefault())
	return client.ReqRaw(http.MethodDelete, url, postbody, args...)
}

// GETBind is bind struct with Get method
func GETBind(v any, url string, args ...map[string]string) error {
	resp := GETRaw(url, args...)
	if !resp.OK() && resp.GetError() != nil {
		return resp.GetError()
	}
	if err := resp.GetStruct(&v); err != nil {
		return err
	}
	return nil

}

// POSTBind is bind struct with Get method
func POSTBind(v any, url string, postbody any, args ...map[string]string) error {
	resp := POSTRaw(url, postbody, args...)
	if !resp.OK() && resp.GetError() != nil {
		return resp.GetError()
	}
	if err := resp.GetStruct(&v); err != nil {
		return err
	}
	return nil
}

// POSTFormBind is bind struct with Get method
func POSTFormBind(v any, url string, postbody any, args ...map[string]string) error {
	resp := POSTForm(url, postbody, args...)
	if !resp.OK() && resp.GetError() != nil {
		return resp.GetError()
	}
	if err := resp.GetStruct(&v); err != nil {
		return err
	}
	return nil
}

func POSTForm(url string, postbody any, args ...map[string]string) response.Response {
	query, header := request.Getqueryheader(args...)
	r, err := request.NewReuqest(
		http.MethodPost,
		url,
		request.WithBody(postbody),
		request.WithQuery(query),
		request.WithHeader(header),
		request.WithContentType(config.FormContectType),
	)
	if err != nil {
		return response.Response{Resp: nil, Body: nil, Err: err}
	}
	client := client.NewClient(client.WithDefault())
	resp, err := client.Do(r)
	if err != nil {
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	resp.Body.Close()
	return response.Response{Resp: resp, Body: body, Err: nil}

}

// POSTBinaryBody is binary body upload
func POSTBinaryBody(url string, binfile io.Reader, timeout time.Duration, args ...map[string]string) response.Response {
	query, header := request.Getqueryheader(args...)

	r, err := request.NewReuqest(
		http.MethodPost,
		url,
		request.WithContentType("binary/octet-stream"),
		request.WithQuery(query),
		request.WithHeader(header),
	)
	if err != nil {
		return response.Response{Resp: nil, Body: nil, Err: err}
	}
	client := client.NewClient(client.WithTimeOut(timeout))

	resp, err := client.Do(r)
	if err != nil {
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	resp.Body.Close()
	return response.Response{Resp: resp, Body: body, Err: nil}

}

// POSTMultiPartUpload is upload file , files key is fieldname of file ,file name is in fields key
func POSTMultiPartUpload(url string, files map[string]io.Reader, fields map[string]string, timeout time.Duration, args ...map[string]string) response.Response {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	for name, file := range files {
		filename, ok := fields[name]
		if !ok {
			return response.Response{Resp: nil, Body: nil, Err: errors.New(name + "is not found in the fields")}
		}
		writer, err := writer.CreateFormFile(name, filename)
		if err != nil {
			return response.Response{Resp: nil, Body: nil, Err: err}
		}
		_, err = io.Copy(writer, file)
		if err != nil {
			return response.Response{Resp: nil, Body: nil, Err: err}
		}
		delete(fields, name)
	}
	for k, v := range fields {
		err := writer.WriteField(k, v)
		if err != nil {
			return response.Response{Resp: nil, Body: nil, Err: err}
		}
	}
	err := writer.Close()
	if err != nil {
		return response.Response{Resp: nil, Body: nil, Err: err}
	}
	query, header := request.Getqueryheader(args...)
	r, err := request.NewReuqest(
		http.MethodPost,
		url,
		request.WithContentType(writer.FormDataContentType()),
		request.WithQuery(query),
		request.WithHeader(header),
	)
	if err != nil {
		return response.Response{Resp: nil, Body: nil, Err: err}
	}
	client := client.NewClient(client.WithTimeOut(timeout))
	resp, err := client.Do(r)
	if err != nil {
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response.Response{Resp: resp, Body: nil, Err: err}
	}
	resp.Body.Close()
	return response.Response{Resp: resp, Body: body, Err: nil}
}
