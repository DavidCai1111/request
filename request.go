package request

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/go-http-utils/headers"
)

// Version is this package's version number.
const Version = "0.0.1"

// Client is a HTTP client.
type Client struct {
	c           *http.Client
	req         *http.Request
	ctx         *context.Context
	queryValues url.Values
	err         error
}

// New returns a Client.
func New(c *http.Client) *Client {
	if c == nil {
		c = &http.Client{}
	}

	return &Client{c: c, req: &http.Request{}}
}

func (c *Client) Request(method string, URL string) *Client {
	if c.err != nil {
		return c
	}

	c.req.Method = method
	u, err := url.Parse(URL)

	if err != nil {
		c.err = err
		return c
	}

	c.req.URL = u

	return c
}

func (c *Client) Get(URL string) *Client {
	return c.Request(http.MethodGet, URL)
}

func (c *Client) Post(URL string) *Client {
	return c.Request(http.MethodPost, URL)
}

func (c *Client) Head(URL string) *Client {
	return c.Request(http.MethodHead, URL)
}

func (c *Client) Delete(URL string) *Client {
	return c.Request(http.MethodDelete, URL)
}

func (c *Client) Set(key, value string) *Client {
	c.req.Header.Set(key, value)

	return c
}

func (c *Client) Add(key, value string) *Client {
	c.req.Header.Add(key, value)

	return c
}

func (c *Client) Type(t string) *Client {
	return c.Set(headers.ContentType, t)
}

func (c *Client) Accept(t string) *Client {
	return c.Set(headers.Accept, t)
}

func (c *Client) Query(queryValues url.Values) *Client {
	return c
}

func (c *Client) Cookie(cookie http.Cookie) *Client {
	return c
}

func (c *Client) SortQuery() *Client {
	return c
}

func (c *Client) Timeout(timeout time.Duration) *Client {
	return c
}

func (c *Client) Redirects(count int) *Client {
	return c
}

func (c *Client) Auth(name, password string) *Client {
	return c
}

func (c *Client) WithCredentials() *Client {
	return c
}

func (c *Client) Field() *Client {
	return c
}

func (c *Client) Attach() *Client {
	return c
}

func (c *Client) End() (*Response, error) {
	if c.err != nil {
		return nil, c.err
	}

	res, err := c.c.Do(c.req)

	if err != nil {
		return nil, err
	}

	return &Response{Response: res}, nil
}
