package request

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/go-http-utils/headers"
)

// Response represents a HTTP response.
type Response struct {
	*http.Response

	raw *bytes.Buffer
}

// Raw returns the raw body of the response.
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
	raw, err := r.Raw()

	if err != nil {
		return nil, err
	}

	var reader io.ReadCloser

	switch r.Header.Get(headers.ContentEncoding) {
	case "gzip":
		if reader, err = gzip.NewReader(r.raw); err != nil {
			return nil, err
		}
	case "deflate":
		if reader, err = zlib.NewReader(r.raw); err != nil {
			return nil, err
		}
	}

	if reader == nil {
		return raw, nil
	}

	defer reader.Close()
	b, err := ioutil.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// JSON returns the reponse body in forms of JSON.
func (r *Response) JSON() (*simplejson.Json, error) {
	b, err := r.Content()

	if err != nil {
		return nil, err
	}

	return simplejson.NewJson(b)
}

// Text returns the reponse body in forms of text string.
func (r *Response) Text() (string, error) {
	b, err := r.Content()

	if err != nil {
		return "", err
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
