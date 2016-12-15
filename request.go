package request

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/go-http-utils/headers"
)

// Version is this package's version number.
const Version = "0.0.1"

// Errors used by this package.
var (
	ErrTimeout            = errors.New("request: request time out")
	ErrExceedMaxRedirects = errors.New("request: exceed max redirects")
	ErrBasicAuthFailed    = errors.New("request: basic auth failed")
	ErrNotPOST            = errors.New("request: method is not POST when use form")
	ErrLackURL            = errors.New("request: lack URL")
	ErrLackMethod         = errors.New("request: lack method")
)

type maxRedirects int

func (mr maxRedirects) check(req *http.Request, via []*http.Request) error {
	if len(via) >= int(mr) {
		return ErrExceedMaxRedirects
	}
	return nil
}

type basicAuthInfo struct {
	name     string
	password string
}

// Client is a HTTP client.
type Client struct {
	cli       *http.Client
	req       *http.Request
	res       *Response
	mw        *multipart.Writer
	mwBuf     *bytes.Buffer
	url       *url.URL
	cookies   []*http.Cookie
	basicAuth *basicAuthInfo
	header    http.Header
	method    string
	vals      url.Values
	timeout   time.Duration
	redirects maxRedirects
	err       error
}

// New returns an new HTTP request Client.
func New(c *http.Client) *Client {
	if c == nil {
		c = &http.Client{}
	}

	return &Client{
		cli:     c,
		header:  http.Header{},
		cookies: []*http.Cookie{},
	}
}

// To defines the HTTP method and URL of this request.
func (c *Client) To(method string, URL string) *Client {
	c.method = method
	u, err := url.Parse(URL)

	if err != nil {
		c.err = err
		return c
	}

	c.url = u
	c.vals = u.Query()

	return c
}

// Get is the shortcut of To("GET", URL) .
func (c *Client) Get(URL string) *Client {
	return c.To(http.MethodGet, URL)
}

// Post is the shortcut of To("POST", URL) .
func (c *Client) Post(URL string) *Client {
	return c.To(http.MethodPost, URL)
}

// Head is the shortcut of To("HEAD", URL) .
func (c *Client) Head(URL string) *Client {
	return c.To(http.MethodHead, URL)
}

// Delete is the shortcut of To("DELETE", URL) .
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

// Type sets the "Content-Type" request header to the given value.
func (c *Client) Type(t string) *Client {
	return c.Set(headers.ContentType, t)
}

// Accept sets the "Accept" request header to the given value.
func (c *Client) Accept(t string) *Client {
	return c.Set(headers.Accept, t)
}

// Query sets the URL query-string to the given value.
func (c *Client) Query(vals url.Values) *Client {
	for k, vs := range vals {
		for _, v := range vs {
			c.vals.Add(k, v)
		}
	}

	return c
}

// Cookie sets the cookie which this request will carry.
func (c *Client) Cookie(cookie *http.Cookie) *Client {
	c.cookies = append(c.cookies, cookie)

	return c
}

// Timeout sets the timeout of this request, if the request
// is timeout, it will return ErrTimeout.
func (c *Client) Timeout(timeout time.Duration) *Client {
	c.timeout = timeout

	return c
}

// Redirects sets the max redirects count for this request.
// If not setted, request will use its default policy,
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
// the "Content-Type" header of this request will be automatically set to
// "application/x-www-form-urlencoded".
func (c *Client) Field(vals url.Values) *Client {
	c.ensureMultiWriter()

	for k, vs := range vals {
		for _, v := range vs {
			if err := c.mw.WriteField(k, v); err != nil {
				c.err = err
				return c
			}
		}
	}

	return c
}

// Attach adds the attached file to the form.
func (c *Client) Attach(name, path, filename string) *Client {
	c.ensureMultiWriter()

	fw, err := c.mw.CreateFormFile(name, filename)

	if err != nil {
		c.err = err
		return c
	}

	file, err := os.Open(path)

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

// End sends the request and get the response of it.
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

	if err := c.assembleReq(); err != nil {
		c.err = err
		return nil, err
	}

	ch := make(chan struct{})

	go func() {
		defer close(ch)
		defer func() { ch <- struct{}{} }()

		response, err := c.cli.Do(c.req)

		if err != nil {
			c.err = err
			return
		}

		c.res = &Response{Response: response}
	}()

	select {
	case <-ch:
	case <-time.After(c.timeout):
		return nil, ErrTimeout
	}

	if c.err != nil {
		return nil, c.err
	}

	return c.res, nil
}

// JSON sends the request and get the JSON of the response.
func (c *Client) JSON() (*simplejson.Json, error) {
	if _, err := c.End(); err != nil {
		return nil, err
	}

	return c.res.JSON()
}

func (c *Client) ensureMultiWriter() {
	if c.mw == nil {
		c.mwBuf = bytes.NewBuffer(nil)
		c.mw = multipart.NewWriter(c.mwBuf)
	}
}

func (c *Client) assembleReq() error {
	c.url.RawQuery = c.vals.Encode()

	req, err := http.NewRequest(c.method, c.url.String(), c.mwBuf)

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
