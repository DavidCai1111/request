# request
[![Build Status](https://travis-ci.org/DavidCai1993/request.svg?branch=master)](https://travis-ci.org/DavidCai1993/request)
[![Coverage Status](https://coveralls.io/repos/github/DavidCai1993/request/badge.svg?branch=master)](https://coveralls.io/github/DavidCai1993/request?branch=master)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/DavidCai1993/request)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/DavidCai1993/request/master/LICENSE)

A concise HTTP request client for Go. It provides elegant and chainalbe API to make you request with happiness.

## Installation

```go
go get -u github.com/DavidCai1993/request
```

## Documentation

API documentation can be found here: https://godoc.org/github.com/DavidCai1993/request

## Usage

### Example:

```go
json, err = request.
  Post("http://mysite.com").
  Timeout(30*time.Second).
  Send(map[string]string{"name": "David"}).
  Set("X-HEADER-KEY", "foo").
  Accept("application/json").
  JSON()
```

### GET:

```go
res, err = request.
  Get("http://mysite.com").
  End()
```

### Cookie:

```go
text, err = request.
  Get("http://mysite.com/get").
  Cookie(&http.Cookie{Name: "name", Value: "David"}).
  Text()
```

### Basic Authentication

```go
json, err = request.
  Get("http://mysite.com/somebooks").
  Auth("name", "passwd").
  JSON()
```

### Form with Attachments

```go
json, err = request.
  Post("http://mysite.com/form").
  Field(url.Values{"key": []string{"value1"}}).
  Attach("test.md", "./README.md", "README.md").
  JSON()
```
