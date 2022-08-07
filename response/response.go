package response

import "net/http"

// Response is  interface TODO:
type Response interface {
	ParseBody(*http.Response) ([]byte, error)
	Error() error
	GETBind(any) error
	POSTBind(any) error
	PUTBind(any) error
	PATCHBind(any) error
	DELETEBind(any) error
	String() (string, error)
}
