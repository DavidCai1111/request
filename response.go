package request

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-http-utils/headers"
)

// GetIndex searches value from []interface{} by index
func GetIndex(v interface{}, index int) interface{} {
	switch v.(type) {
	case []interface{}:
		res := v.([]interface{})
		if len(res) > index {
			return res[index]
		}
		return nil
	case *[]interface{}:
		res := v.(*[]interface{})
		return GetIndex(*res, index)
	default:
		return nil
	}
}

// GetPath searches value from map[string]interface{} by path
func GetPath(v interface{}, branch ...string) interface{} {
	switch v.(type) {
	case map[string]interface{}:
		res := v.(map[string]interface{})
		switch len(branch) {
		case 0:
			return nil // should return nil when no branch
		case 1:
			return res[branch[0]]
		default:
			return GetPath(res[branch[0]], branch[1:]...)
		}
	case *map[string]interface{}:
		res := v.(*map[string]interface{})
		return GetPath(*res, branch...)
	default:
		return nil
	}
}

// Response represents the response from a HTTP request.
type Response struct {
	*http.Response

	raw     *bytes.Buffer
	content []byte
}

// Raw returns the raw bytes body of the response.
func (r *Response) Raw() ([]byte, error) {
	if r.raw != nil {
		return r.raw.Bytes(), nil
	}

	b, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		return nil, err
	}

	r.raw = bytes.NewBuffer(b)

	return b, nil
}

// Content returns the content of the response body, it will handle
// the compression.
func (r *Response) Content() ([]byte, error) {
	if r.content != nil {
		return r.content, nil
	}

	rawBytes, err := r.Raw()

	if err != nil {
		return nil, err
	}

	var reader io.ReadCloser

	switch r.Header.Get(headers.ContentEncoding) {
	case "gzip":
		if reader, err = gzip.NewReader(bytes.NewBuffer(r.raw.Bytes())); err != nil {
			return nil, err
		}
	case "deflate":
		reader = flate.NewReader(bytes.NewBuffer(r.raw.Bytes()))
	}

	if reader == nil {
		r.content = rawBytes

		return rawBytes, nil
	}

	defer reader.Close()
	b, err := ioutil.ReadAll(reader)

	// If gzip or deflate decoding failed, try zlib decoding instead.
	// The body may be wrapped in the zlib data format.
	if err != nil {
		var zlibReader io.ReadCloser

		if zlibReader, err = zlib.NewReader(bytes.NewBuffer(r.raw.Bytes())); err != nil {
			return nil, err
		}
		defer zlibReader.Close()

		if b, err = ioutil.ReadAll(zlibReader); err != nil {
			return nil, err
		}
	}

	r.content = b

	return b, nil
}

// JSON returns the reponse body with JSON format.
func (r *Response) JSON(v ...interface{}) (interface{}, error) {
	b, err := r.Content()
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(r.Header.Get(headers.ContentType), "application/json") {
		err := r.Status
		if len(b) > 0 {
			err = string(b)
		}
		return nil, errors.New(err)
	}

	var res interface{}
	if len(v) > 0 {
		res = v[0]
	} else {
		res = new(map[string]interface{})
	}

	if err = json.Unmarshal(b, res); err != nil {
		return nil, err
	}

	if !r.OK() {
		return res, ErrStatusNotOk
	}

	return res, nil
}

// Text returns the reponse body with text format.
func (r *Response) Text() (string, error) {
	b, err := r.Content()

	if err != nil {
		return "", err
	}

	if !r.OK() {
		return string(b), ErrStatusNotOk
	}

	return string(b), nil
}

// URL returns url of the final request.
func (r *Response) URL() (*url.URL, error) {
	u := r.Request.URL

	if r.StatusCode == http.StatusMovedPermanently ||
		r.StatusCode == http.StatusFound ||
		r.StatusCode == http.StatusSeeOther ||
		r.StatusCode == http.StatusTemporaryRedirect {
		location, err := r.Location()

		if err != nil {
			return nil, err
		}

		u = u.ResolveReference(location)
	}

	return u, nil
}

// Reason returns the status text of the response status code.
func (r *Response) Reason() string {
	return http.StatusText(r.StatusCode)
}

// OK returns whether the reponse status code is less than 400.
func (r *Response) OK() bool {
	return r.StatusCode < 400
}
