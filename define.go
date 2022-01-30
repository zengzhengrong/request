package request

import (
	"net/http"
	"os"
	"time"
)

const (
	DefaultMultipartMemory             = 32 << 20 // 32 MB
	DefaultContentType                 = "application/json"
	JsonContectType                    = DefaultContentType
	FormContectType                    = "application/x-www-form-urlencoded"
	DefaultTimeout                     = 60 * time.Second
	DefaultTLSConfigInsecureSkipVerify = true
	PiplineCtxValueKey                 = "values"
)

func SetDefaultDebug() bool {
	var BoolFlagMap = map[string]bool{
		"1":     true,
		"true":  true,
		"True":  true,
		"0":     false,
		"false": false,
		"False": false,
	}

	debug := os.Getenv("REQUEST_DEBUG")
	bl, ok := BoolFlagMap[debug]
	if !ok {
		return false
	}
	return bl
}

var MaxUploadThreads int = 20
var DefaultDebug = SetDefaultDebug
var DefaultCheckRedirect = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}
