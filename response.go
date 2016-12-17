package request

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/go-http-utils/headers"
)

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
func (r *Response) JSON() (*simplejson.Json, error) {
	if !r.OK() {
		return nil, ErrStatusNotOk
	}

	b, err := r.Content()

	if err != nil {
		return nil, err
	}

	return simplejson.NewJson(b)
}

// Text returns the reponse body with text format.
func (r *Response) Text() (string, error) {
	if !r.OK() {
		return "", ErrStatusNotOk
	}

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

// OK returns whether the reponse status code is less than 400.
func (r *Response) OK() bool {
	return r.StatusCode < 400
}
