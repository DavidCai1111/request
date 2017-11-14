package request

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/suite"
)

const testHost = "http://httpbin.org"

type RequestSuite struct {
	suite.Suite

	c *Client
}

func (s *RequestSuite) SetupTest() {
	s.c = New()
}

func (s *RequestSuite) TestGet() {
	res, err := s.c.Get(testHost).End()

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
}

func (s *RequestSuite) TestQuickGet() {
	res, err := Get(testHost).End()

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
}

func (s *RequestSuite) TestPost() {
	body := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}

	j, err := s.c.
		Post(testHost + "/post").
		Send(body).
		JSON()

	s.Nil(err)
	s.Equal(body["k1"], GetPath(j, "json", "k1").(string))
	s.Equal(body["k2"], GetPath(j, "json", "k2").(string))
}

func (s *RequestSuite) TestPostWithString() {
	body := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}

	j, err := s.c.
		Post(testHost + "/post").
		Send(`{"k1":"v1","k2":"v2"}`).
		JSON()

	s.Nil(err)
	s.Equal(body["k1"], GetPath(j, "json", "k1").(string))
	s.Equal(body["k2"], GetPath(j, "json", "k2").(string))
}

func (s *RequestSuite) TestQuickPost() {
	body := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}

	type result struct {
		Body map[string]string `json:"json"`
	}

	j, err := Post(testHost + "/post").
		Send(body).
		JSON(&result{})

	s.Nil(err)
	s.Equal(body["k1"], j.(*result).Body["k1"])
	s.Equal(body["k2"], j.(*result).Body["k2"])
}

func (s *RequestSuite) TestPut() {
	body := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}

	j, err := s.c.
		Put(testHost + "/put").
		Send(body).
		JSON()

	s.Nil(err)
	s.Equal(body["k1"], GetPath(j, "json", "k1").(string))
	s.Equal(body["k2"], GetPath(j, "json", "k2").(string))
}

func (s *RequestSuite) TestQuickPut() {
	body := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}

	j, err := Put(testHost + "/put").
		Send(body).
		JSON()

	s.Nil(err)
	s.Equal(body["k1"], GetPath(j, "json", "k1").(string))
	s.Equal(body["k2"], GetPath(j, "json", "k2").(string))
}

func (s *RequestSuite) TestDelete() {
	res, err := s.c.
		Delete(testHost + "/delete?k1=v1&k2=v2").
		End()

	s.Nil(err)
	s.True(res.OK())
}

func (s *RequestSuite) TestQuickDelete() {
	res, err := Delete(testHost + "/delete?k1=v1&k2=v2").End()

	s.Nil(err)
	s.True(res.OK())
}

func (s *RequestSuite) TestToErr() {
	s.c.err = fmt.Errorf("test")

	_, err := s.c.
		Get(testHost + "/get").
		End()

	s.EqualError(err, "test")
}

func (s *RequestSuite) TestToWrongURL() {
	_, err := s.c.
		To(http.MethodGet, "%").
		End()

	s.NotNil(err)
}

func (s *RequestSuite) TestSet() {
	j, err := s.c.
		Get(testHost+"/headers").
		Set("X-Test-Key", "X-TEST-VALUE").
		JSON()

	s.Nil(err)
	s.Equal("X-TEST-VALUE", GetPath(j, "headers", "X-Test-Key").(string))
}

func (s *RequestSuite) TestAdd() {
	j, err := s.c.
		Get(testHost+"/headers").
		Add("X-Test-Key", "X-TEST-VALUE1").
		// Add("X-Test-Key", "X-TEST-VALUE2").
		JSON()

	s.Nil(err)
	s.Equal("X-TEST-VALUE1", GetPath(j, "headers", "X-Test-Key").(string))
}

func (s *RequestSuite) TestHeader() {
	h := http.Header{
		"X-Test-Key1": []string{"X-TEST-VALUE1"},
		"X-Test-Key2": []string{"X-TEST-VALUE2"},
	}

	j, err := s.c.
		Get(testHost+"/headers").
		Set("X-Test-Key1", "X-TEST-VALUE3").
		Set("X-Test-Key3", "X-TEST-VALUE4").
		Header(h).
		JSON()

	s.Nil(err)
	s.Equal("X-TEST-VALUE1", GetPath(j, "headers", "X-Test-Key1").(string))
	s.Equal("X-TEST-VALUE2", GetPath(j, "headers", "X-Test-Key2").(string))
	s.Equal("X-TEST-VALUE4", GetPath(j, "headers", "X-Test-Key3").(string))
}

func (s *RequestSuite) TestType() {
	j, err := s.c.
		Get(testHost + "/headers").
		Type("text/plain+test").
		JSON()

	s.Nil(err)
	s.Equal("text/plain+test", GetPath(j, "headers", headers.ContentType).(string))
}

func (s *RequestSuite) TestAccept() {
	j, err := s.c.
		Get(testHost + "/headers").
		Accept("text/plain+test").
		JSON()

	s.Nil(err)
	s.Equal("text/plain+test", GetPath(j, "headers", headers.Accept).(string))
}

func (s *RequestSuite) TestPredefinedAccept() {
	j, err := s.c.
		Get(testHost + "/headers").
		Accept("text").
		JSON()

	s.Nil(err)
	s.Equal("text/plain", GetPath(j, "headers", headers.Accept).(string))
}

func (s *RequestSuite) TestPredefinedType() {
	j, err := s.c.
		Get(testHost + "/headers").
		Type("HTML").
		JSON()

	s.Nil(err)
	s.Equal("text/html", GetPath(j, "headers", headers.ContentType).(string))
}

func (s *RequestSuite) TestTextEmptyURL() {
	_, err := s.c.Text()

	s.Equal(ErrLackURL, err)
}

func (s *RequestSuite) TestPathQuery() {
	j, err := s.c.
		Get(testHost + "/get?a=b").
		JSON()

	s.Nil(err)
	s.Equal("b", GetPath(j, "args", "a").(string))
}

func (s *RequestSuite) TestQuery() {
	v := url.Values{
		"k1": []string{"v1", "v2"},
		"k2": []string{"v3"},
	}

	j, err := s.c.
		Get(testHost + "/get").
		Query(v).
		JSON()

	s.Nil(err)
	s.Equal("v1", GetIndex(GetPath(j, "args", "k1"), 0).(string))
	s.Equal("v2", GetIndex(GetPath(j, "args", "k1"), 1).(string))
	s.Equal("v3", GetPath(j, "args", "k2").(string))
}

func (s *RequestSuite) TestSend() {
	b := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}

	j, err := s.c.
		Post(testHost + "/post").
		Send(b).
		JSON()

	s.Nil(err)
	s.Equal("v1", GetPath(j, "json", "k1").(string))
	s.Equal("v2", GetPath(j, "json", "k2").(string))
}

func (s *RequestSuite) TestSendAfterAttach() {
	_, err := s.c.
		Post(testHost+"/post").
		Attach("test.md", "./README.md", "README.md").
		Send(nil).
		JSON()

	s.Equal(ErrBodyAlreadySet, err)
}

func (s *RequestSuite) TestSendCanNotMarshel() {
	_, err := s.c.
		Post(testHost + "/post").
		Send(make(chan bool)).
		JSON()

	s.NotNil(err)
}

func (s *RequestSuite) TestCookie() {
	cookie := &http.Cookie{
		Name:  "k1",
		Value: "v1",
	}

	j, err := s.c.
		Get(testHost + "/cookies").
		Cookie(cookie).
		JSON()

	s.Nil(err)
	s.Equal("v1", GetPath(j, "cookies", "k1").(string))
}

func (s *RequestSuite) TestCookieJar() {
	cookie := &http.Cookie{
		Name:  "k1",
		Value: "v1",
	}

	cookieJar, err := cookiejar.New(nil)

	s.Nil(err)

	u, err := url.Parse(testHost + "/cookies")

	s.Nil(err)

	cookieJar.SetCookies(u, []*http.Cookie{cookie})

	j, err := s.c.
		Get(testHost + "/cookies").
		CookieJar(cookieJar).
		JSON()

	s.Nil(err)
	s.Equal("v1", GetPath(j, "cookies", "k1").(string))
}

func (s *RequestSuite) TestTimeout() {
	_, err := s.c.
		Get(testHost + "/get").
		Timeout(1 * time.Millisecond).
		JSON()

	s.NotNil(err)
}

func (s *RequestSuite) TestRedirect() {
	_, err := s.c.
		Get(testHost + "/redirect/5").
		Redirects(4).
		JSON()

	s.NotNil(err)
}

func (s *RequestSuite) TestAuth() {
	j, err := s.c.
		Get(testHost+"/basic-auth/user/passwd").
		Auth("user", "passwd").
		JSON()

	s.Nil(err)
	s.Equal(true, GetPath(j, "authenticated").(bool))
	s.Equal("user", GetPath(j, "user").(string))
}

func (s *RequestSuite) TestNotAuthPass() {
	res, err := s.c.
		Get(testHost+"/basic-auth/user/passwd").
		Auth("user", "passwd1").
		End()

	s.Nil(err)
	s.Equal(http.StatusUnauthorized, res.StatusCode)
}

func (s *RequestSuite) TestField() {
	v := url.Values{
		"k1": []string{"v1", "v2"},
		"k2": []string{"v3"},
	}

	j, err := s.c.
		Post(testHost + "/post").
		Field(v).
		JSON()

	s.Nil(err)
	s.Equal("v1", GetIndex(GetPath(j, "form", "k1"), 0).(string))
	s.Equal("v2", GetIndex(GetPath(j, "form", "k1"), 1).(string))
	s.Equal("v3", GetPath(j, "form", "k2").(string))
}

func (s *RequestSuite) TestAttach() {
	j, err := s.c.
		Post(testHost+"/post").
		Attach("test.md", "./README.md", "README.md").
		JSON()

	s.Nil(err)
	s.NotEmpty(GetPath(j, "files", "test.md").(string))
}

func (s *RequestSuite) TestAttachAfterSend() {
	_, err := s.c.
		Post(testHost+"/post").
		Send(true).
		Attach("test.md", "./README.md", "README.md").
		JSON()

	s.Equal(ErrBodyAlreadySet, err)
}

func (s *RequestSuite) TestAttachCanNotOpenFile() {
	_, err := s.c.
		Post(testHost+"/post").
		Attach("test.md", "./not-exists.md", "README.md").
		JSON()

	s.NotNil(err)
}

func (s *RequestSuite) TestFieldsAndAttach() {
	v := url.Values{
		"k1": []string{"v1", "v2"},
		"k2": []string{"v3"},
	}

	j, err := s.c.
		Post(testHost+"/post").
		Field(v).
		Attach("test.md", "./README.md", "README.md").
		JSON()

	s.Nil(err)
	s.Equal("v1", GetIndex(GetPath(j, "form", "k1"), 0).(string))
	s.Equal("v2", GetIndex(GetPath(j, "form", "k1"), 1).(string))
	s.Equal("v3", GetPath(j, "form", "k2").(string))
	s.NotEmpty(GetPath(j, "files", "test.md").(string))
}

func (s *RequestSuite) TestEndWithoutURL() {
	_, err := s.c.End()

	s.Equal(ErrLackURL, err)
}

func (s *RequestSuite) TestEndWithoutMethod() {
	u, err := url.Parse(testHost)

	s.Nil(err)
	s.c.url = u
	_, err = s.c.End()
	s.Equal(ErrLackMethod, err)
}

func (s *RequestSuite) TestProxy() {
	proxy := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		res.Write([]byte("{\"proxy\": \"test\"}"))
	}))

	defer proxy.Close()

	j, err := s.c.
		Post(testHost + "/get").
		Proxy(proxy.URL).
		JSON()

	s.Nil(err)
	s.Equal("test", GetPath(j, "proxy").(string))
}

func (s *RequestSuite) TestProxyInvalidURL() {
	_, err := s.c.
		Post(testHost + "/get").
		Proxy("%%%%").
		JSON()

	s.NotNil(err)
}

func (s *RequestSuite) TestProxyInvalidSocks5URL() {
	_, err := s.c.
		Post(testHost + "/get").
		Proxy("socks5://").
		JSON()

	s.NotNil(err)
}

func (s *RequestSuite) TestReq() {
	req, err := s.c.Post(testHost + "/post").
		Accept("json").
		Req()

	s.Nil(err)
	s.Equal(req.Header.Get(headers.Accept), "application/json")
}

func TestRequest(t *testing.T) {
	suite.Run(t, new(RequestSuite))
}
