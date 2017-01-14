package request_test

import (
	"net/http"
	"net/url"
	"time"

	"github.com/DavidCai1993/request"
)

var (
	text string
	err  error
	res  *request.Response
	json interface{}
)

func Example() {
	res, err = request.
		Get("http://mysite.com").
		End()

	// json has the type *simplejson.Json
	json, err = request.
		Post("http://mysite.com").
		Timeout(30*time.Second).
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

func ExampleGetWithStruct() {
	type MyResult struct {
		Code  int                    `json:"code"`
		Error string                 `json:"error"`
		Data  map[string]interface{} `json:"data"`
	}

	json, err = request.
		Get("http://mysite.com").
		JSON(new(MyResult))
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

func ExampleClient_Proxy() {
	json, err = request.
		Get("http://mysite.com/somebooks").
		Proxy("http://myproxy.com:8080").
		JSON()
}

func ExampleClient_Attach() {
	json, err = request.
		Post("http://mysite.com/readme").
		Field(url.Values{"key": []string{"value1"}}).
		Attach("test.md", "./README.md", "README.md").
		JSON()
}
