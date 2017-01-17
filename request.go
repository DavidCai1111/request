package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-http-utils/headers"
	"golang.org/x/net/proxy"
)

// Version is this package's version number.
const Version = "1.5.1"

// Errors used by this package.
var (
	ErrNotPOST        = errors.New("request: method is not POST when using form")
	ErrLackURL        = errors.New("request: request lacks URL")
	ErrLackMethod     = errors.New("request: request lacks method")
	ErrBodyAlreadySet = errors.New("request: request body has already been set")
	ErrStatusNotOk    = errors.New("request: status code is not ok (>= 400)")
)

type maxRedirects int

func (mr maxRedirects) check(req *http.Request, via []*http.Request) error {
	if len(via) >= int(mr) {
		return fmt.Errorf("request: exceed max redirects")
	}
	return nil
}

type basicAuthInfo struct {
	name     string
	password string
}

// Client is a HTTP client which provides usable and chainable methods.
type Client struct {
	cli       *http.Client
	req       *http.Request
	res       *Response
	method    string
	url       *url.URL
	queryVals url.Values
	formVals  url.Values
	mw        *multipart.Writer
	mwBuf     *bytes.Buffer
	body      io.Reader
	basicAuth *basicAuthInfo
	header    http.Header
	cookies   []*http.Cookie
	timeout   time.Duration
	redirects maxRedirects
	err       error
}

// New returns a new instance of Client.
func New() *Client {
	c := &Client{
		cli:      new(http.Client),
		header:   make(http.Header),
		formVals: make(url.Values),
		cookies:  make([]*http.Cookie, 0),
		mwBuf:    bytes.NewBuffer(nil),
	}

	c.mw = multipart.NewWriter(c.mwBuf)

	return c
}

// To defines the method and URL of the request.
func (c *Client) To(method string, URL string) *Client {
	c.method = method
	u, err := url.Parse(URL)

	if err != nil {
		c.err = err
		return c
	}

	c.url = u
	c.queryVals = u.Query()

	return c
}

// Get equals New().Get(URL) to let you start a GET request conveniently.
func Get(URL string) *Client {
	return New().Get(URL)
}

// Post equals New().Post(URL) to let you start a POST request conveniently.
func Post(URL string) *Client {
	return New().Post(URL)
}

// Put equals New().Put(URL) to let you start a PUT request conveniently.
func Put(URL string) *Client {
	return New().Put(URL)
}

// Delete equals New().Delete(URL) to let you start a DELETE request
// conveniently.
func Delete(URL string) *Client {
	return New().Delete(URL)
}

// Get equals To("GET", URL) .
func (c *Client) Get(URL string) *Client {
	return c.To(http.MethodGet, URL)
}

// Post equals To("POST", URL) .
func (c *Client) Post(URL string) *Client {
	return c.To(http.MethodPost, URL)
}

// Put equals To("PUT", URL) .
func (c *Client) Put(URL string) *Client {
	return c.To(http.MethodPut, URL)
}

// Delete equals To("DELETE", URL) .
func (c *Client) Delete(URL string) *Client {
	return c.To(http.MethodDelete, URL)
}

// Set sets the request header entries associated with key to the single
// element value. It replaces any existing values associated with key.
func (c *Client) Set(key, value string) *Client {
	c.header.Set(key, value)

	return c
}

// Add adds the key, value pair to the request header.It appends to any
// existing values associated with key.
func (c *Client) Add(key, value string) *Client {
	c.header.Add(key, value)

	return c
}

// Header sets all key, value pairs in h to the request header, it replaces any
// existing values associated with key.
func (c *Client) Header(h http.Header) *Client {
	for k, v := range h {
		c.header[k] = v
	}

	return c
}

var typesMap = map[string]string{
	"html":       "text/html",
	"json":       "application/json",
	"xml":        "application/xml",
	"text":       "text/plain",
	"urlencoded": "application/x-www-form-urlencoded",
	"form":       "application/x-www-form-urlencoded",
	"form-data":  "application/x-www-form-urlencoded",
	"multipart":  "multipart/form-data",
}

// Type sets the "Content-Type" request header to the given value.
func (c *Client) Type(t string) *Client {
	if typ, ok := typesMap[strings.ToLower(t)]; ok {
		return c.Set(headers.ContentType, typ)
	}

	return c.Set(headers.ContentType, t)
}

// Accept sets the "Accept" request header to the given value.
func (c *Client) Accept(t string) *Client {
	if typ, ok := typesMap[strings.ToLower(t)]; ok {
		return c.Set(headers.Accept, typ)
	}

	return c.Set(headers.Accept, t)
}

// Query adds the the given value to request's URL query-string.
func (c *Client) Query(vals url.Values) *Client {
	for k, vs := range vals {
		for _, v := range vs {
			c.queryVals.Add(k, v)
		}
	}

	return c
}

// Send sends the body in JSON format, body can be anything which can be
// Marshaled or just Marshaled JSON string.
func (c *Client) Send(body interface{}) *Client {
	if c.body != nil || c.mwBuf.Len() != 0 {
		c.err = ErrBodyAlreadySet
		return c
	}

	switch body := body.(type) {
	case string:
		c.body = bytes.NewBufferString(body)
	default:
		j, err := json.Marshal(body)

		if err != nil {
			c.err = err
			return c
		}

		c.body = bytes.NewReader(j)
	}

	c.Set(headers.ContentType, "application/json")
	return c
}

// Cookie adds the cookie to the request.
func (c *Client) Cookie(cookie *http.Cookie) *Client {
	c.cookies = append(c.cookies, cookie)

	return c
}

// CookieJar adds all cookies in the cookie jar to the request.
func (c *Client) CookieJar(jar http.CookieJar) *Client {
	for _, cookie := range jar.Cookies(c.url) {
		c.Cookie(cookie)
	}

	return c
}

// Timeout specifies a time limit for the request.
// The timeout includes connection time, any
// redirects, and reading the response body. The timer remains
// running after Get, Head, Post, or End return and will
// interrupt reading of the response body.
func (c *Client) Timeout(timeout time.Duration) *Client {
	c.cli.Timeout = timeout

	return c
}

// Redirects sets the max redirects count for the request.
// If not set, request will use its default policy,
// which is to stop after 10 consecutive requests.
func (c *Client) Redirects(count int) *Client {
	c.cli.CheckRedirect = maxRedirects(count).check

	return c
}

// Auth sets the request's Authorization header to use HTTP Basic
// Authentication with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password are not
// encrypted.
func (c *Client) Auth(name, password string) *Client {
	c.basicAuth = &basicAuthInfo{name: name, password: password}

	return c
}

// Field sets the field values like form fields in HTML. Once it was set,
// the "Content-Type" header of the request will be automatically set to
// "application/x-www-form-urlencoded".
func (c *Client) Field(vals url.Values) *Client {
	for k, vs := range vals {
		for _, v := range vs {
			c.formVals.Add(k, v)
		}
	}

	c.Type("application/x-www-form-urlencoded")
	return c
}

// Attach adds the attachment file to the form. Once the attachment was
// set, the "Content-Type" will be set to "multipart/form-data; boundary=xxx"
// automatically.
func (c *Client) Attach(fieldname, path, filename string) *Client {
	if c.body != nil {
		c.err = ErrBodyAlreadySet
		return c
	}

	file, err := os.Open(path)

	if err != nil {
		c.err = err
		return c
	}

	fw, err := c.mw.CreateFormFile(fieldname, filename)

	if err != nil {
		c.err = err
		return c
	}

	if _, err = io.Copy(fw, file); err != nil {
		c.err = err
		return c
	}

	return c
}

// Proxy sets the address of the proxy which used by the request.
func (c *Client) Proxy(addr string) *Client {
	u, err := url.Parse(addr)

	if err != nil {
		c.err = err
		return c
	}

	switch u.Scheme {
	case "http", "https":
		c.cli.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	case "socks5":
		dialer, err := proxy.FromURL(u, proxy.Direct)

		if err != nil {
			c.err = err
			return c
		}

		c.cli.Transport = &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			Dial:                dialer.Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}

	return c
}

// End sends the HTTP request and returns the HTTP reponse.
//
// An error is returned if caused by client policy (such as timeout), or
// failure to speak HTTP (such as a network connectivity problem), or generated
// by former chained methods. A non-2xx status code doesn't cause an error.
func (c *Client) End() (*Response, error) {
	if c.url == nil {
		return nil, ErrLackURL
	}

	if c.method == "" {
		return nil, ErrLackMethod
	}

	if c.err != nil || c.res != nil {
		return c.res, c.err
	}

	if err := c.assemble(); err != nil {
		c.err = err
		return nil, err
	}

	response, err := c.cli.Do(c.req)

	if err != nil {
		c.err = err
		return nil, err
	}

	c.res = &Response{Response: response}

	return c.res, nil
}

// JSON sends the HTTP request and returns the reponse body with JSON format.
func (c *Client) JSON(v ...interface{}) (interface{}, error) {
	if _, err := c.End(); err != nil {
		return nil, err
	}

	return c.res.JSON(v...)
}

// Text sends the HTTP request and returns the reponse body with text format.
func (c *Client) Text() (string, error) {
	if _, err := c.End(); err != nil {
		return "", err
	}

	return c.res.Text()
}

func (c *Client) assemble() error {
	c.url.RawQuery = c.queryVals.Encode()

	var buf io.Reader

	if c.mwBuf.Len() != 0 {
		if c.formVals != nil {
			for k, vs := range c.formVals {
				for _, v := range vs {
					if err := c.mw.WriteField(k, v); err != nil {
						return err
					}
				}
			}
		}

		buf = c.mwBuf
		c.Type(c.mw.FormDataContentType())
		c.mw.Close()
	} else if c.formVals != nil && c.body == nil {
		buf = strings.NewReader(c.formVals.Encode())
	} else {
		buf = c.body
	}

	req, err := http.NewRequest(c.method, c.url.String(), buf)

	if err != nil {
		return err
	}

	c.req = req
	c.req.Header = c.header

	if c.basicAuth != nil {
		c.req.SetBasicAuth(c.basicAuth.name, c.basicAuth.password)
	}

	for _, cookie := range c.cookies {
		c.req.AddCookie(cookie)
	}

	return nil
}
