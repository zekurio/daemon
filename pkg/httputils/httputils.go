package httputils

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/valyala/fasthttp"
)

// Request executes a HTTP request with the passed method, url, headers and body.
// It returns the response and errors if one occurred.
func Request(method, url string, headers map[string]string, data interface{}) (res *Response, err error) {
	defer func() {
		if err != nil && res != nil {
			res.Release()
		}
	}()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res = responsePool.Get().(*Response)

	req.Header.SetMethod(method)
	req.SetRequestURI(url)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if data != nil {
		err = json.NewEncoder(req.BodyWriter()).Encode(data)
		if err != nil {
			return
		}
	}

	err = fasthttp.Do(req, res.Response)
	return
}

// Get is a shortcut for Request with method GET.
func Get(url string, headers map[string]string) (res *Response, err error) {
	return Request("GET", url, headers, nil)
}

// Post is a shortcut for Request with method POST.
func Post(url string, headers map[string]string, data interface{}) (res *Response, err error) {
	return Request("POST", url, headers, data)
}

// GetFile is a shortcut for Request with method GET and returns the response body as io.Reader file.
func GetFile(url string, headers map[string]string) (file io.Reader, contentType string, err error) {
	resp, err := Get(url, headers)
	if err != nil {
		return
	}
	defer resp.Release()
	file = bytes.NewBuffer(resp.Body())
	contentType = string(resp.Header.Peek("content-type"))
	return
}
