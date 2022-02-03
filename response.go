package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/gjson"
)

type Response struct {
	Resp *http.Response
	Body []byte
	Err  error
}

func (r *Response) statusCodeError() error {
	if r.Resp.ContentLength > 0 {
		return errors.New(string(r.Body))
	}
	return StatusCodeError
}

// OK is StatusCode 200
func (r *Response) OK() bool {
	if r.Err != nil {
		return false
	}
	if r.Resp.StatusCode == 200 {
		return true
	}
	return false
}

func (r *Response) OKByJsonKey(key string, value any) bool {
	if r.Err != nil {
		return false
	}
	res := gjson.GetBytes(r.Body, key)
	if res.Value() == value {
		return true
	}
	return false

}

func (r *Response) Error() string {
	if r.Err != nil {
		err := fmt.Errorf("Response err: %w", r.Err)
		return err.Error()
	}
	return ""
}

func (r *Response) GetError() error {
	if r.Err != nil {
		return fmt.Errorf("Response err: %w", r.Err)
	}
	return nil
}

func (r *Response) GetBody() []byte {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		r.Body, _ = io.ReadAll(r.Resp.Body)
	}
	return r.Body
}

func (r *Response) GetBodyString() string {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		r.Body, _ = io.ReadAll(r.Resp.Body)
	}
	return string(r.Body)
}

func (r *Response) GetString(key string) string {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		r.Body, _ = io.ReadAll(r.Resp.Body)
	}
	return gjson.GetBytes(r.Body, key).String()
}

func (r *Response) GetInt(key string) int64 {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		r.Body, _ = io.ReadAll(r.Resp.Body)
	}
	return gjson.GetBytes(r.Body, key).Int()
}

func (r *Response) GetFloat(key string) float64 {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		r.Body, _ = io.ReadAll(r.Resp.Body)
	}
	return gjson.GetBytes(r.Body, key).Float()
}

func (r *Response) GetMap(key string) map[string]gjson.Result {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		r.Body, _ = io.ReadAll(r.Resp.Body)
	}
	return gjson.GetBytes(r.Body, key).Map()
}

func (r *Response) GetArrary(key string) []gjson.Result {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		r.Body, _ = io.ReadAll(r.Resp.Body)
	}
	return gjson.GetBytes(r.Body, key).Array()
}

func (r *Response) GetStruct(v any) error {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		// use resp body decode
		if err := json.NewDecoder(r.Resp.Body).Decode(v); err != nil {
			return err
		}
		return nil
	}
	body := bytes.NewReader(r.Body)
	if err := json.NewDecoder(body).Decode(v); err != nil {
		return err
	}
	return nil
}

func (r *Response) GetKeyStruct(v any, key string) error {
	if r.Body == nil && r.Err == nil {
		defer r.Resp.Body.Close()
		r.Body, _ = io.ReadAll(r.Resp.Body)
	}
	body := bytes.NewReader([]byte(gjson.GetBytes(r.Body, key).Raw))
	if err := json.NewDecoder(body).Decode(v); err != nil {
		return err
	}
	return nil
}
