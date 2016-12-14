package request

import (
	"bytes"
	"io/ioutil"
	"net/http"

	simplejson "github.com/bitly/go-simplejson"
)

type Response struct {
	*http.Response

	raw *bytes.Buffer
}

func (r Response) Raw() ([]byte, error) {
	// application/x-www-form-urlencoded, application/json, and multipart/form-data
	b, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		return nil, err
	}

	r.raw = bytes.NewBuffer(b)

	return b, nil
}

func (r Response) JSON() (*simplejson.Json, error) {
	b, err := r.Raw()

	if err != nil {
		return nil, err
	}

	return simplejson.NewJson(b)
}

func (r Response) Text() (string, error) {
	b, err := r.Raw()

	if err != nil {
		return "", err
	}

	return string(b), nil
}
