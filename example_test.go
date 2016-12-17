package request_test

import (
	"net/http"

	"net/url"

	"github.com/DavidCai1993/request"
	simplejson "github.com/bitly/go-simplejson"
)

var (
	text string
	err  error
	res  *request.Response
	json *simplejson.Json
)

func Example() {
	res, err = request.
		Get("http://mysite.com").
		End()

	// json has the type *simplejson.Json
	json, err = request.
		Post("http://mysite.com").
		Send(map[string]string{"name": "David"}).
		Set("X-HEADER-KEY", "foo").
		Accept("application/json").
		JSON()
}

func ExampleGet() {
	json, err = request.
		Get("http://mysite.com").
		JSON()
}

func ExamplePost() {
	json, err = request.
		Post("http://mysite.com").
		Send(map[string]string{"name": "David"}).
		Set("X-HEADER-KEY", "foo").
		Accept("application/json").
		JSON()
}

func ExampleClient_Cookie() {
	text, err = request.
		Get("http://mysite.com/get").
		Cookie(&http.Cookie{Name: "name", Value: "David"}).
		Text()
}

func ExampleClient_Auth() {
	json, err = request.
		Get("http://mysite.com/somebooks").
		Auth("name", "passwd").
		JSON()
}

func ExampleClient_Attach() {
	json, err = request.
		Post("http://mysite.com/readme").
		Field(url.Values{"key": []string{"value1"}}).
		Attach("test.md", "./README.md", "README.md").
		JSON()
}
