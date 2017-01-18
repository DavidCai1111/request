package request

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/suite"
)

type ResponseSuite struct {
	suite.Suite

	c *Client
}

func (s *ResponseSuite) SetupTest() {
	s.c = New()
}

func (s *ResponseSuite) TestGet() {
	res, err := s.c.Get(testHost).End()

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
	s.Equal("OK", res.Reason())
}

func (s *ResponseSuite) TestRaw() {
	res, err := s.c.Get(testHost).End()
	s.Nil(err)

	raw1, err := res.Raw()
	s.Nil(err)

	raw2, err := res.Raw()
	s.Nil(err)

	s.Equal(raw1, raw2)
}

func (s *ResponseSuite) TestContent() {
	res, err := s.c.Get(testHost).End()
	s.Nil(err)

	c1, err := res.Content()
	s.Nil(err)

	c2, err := res.Content()
	s.Nil(err)

	s.Equal(c1, c2)
}

func (s *ResponseSuite) TestURL() {
	res, err := s.c.Get(testHost).End()

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)

	u, err := res.URL()

	s.Nil(err)
	s.Equal(testHost, u.String())
}

func (s *ResponseSuite) TestRedirectURL() {
	res, err := s.c.Get(testHost).End()

	res.StatusCode = http.StatusMovedPermanently
	res.Header.Set(headers.Location, "test")

	s.Nil(err)

	u, err := res.URL()

	s.Nil(err)
	s.Equal(testHost+"/test", u.String())
}

func (s *ResponseSuite) TestGzip() {
	j, err := s.c.
		Get(testHost+"/gzip").
		Set(headers.AcceptEncoding, "gzip, deflate").
		JSON()

	s.Nil(err)
	s.Equal(true, GetPath(j, "gzipped").(bool))
}

func (s *ResponseSuite) TestDeflate() {
	j, err := s.c.
		Get(testHost + "/deflate").
		JSON()

	s.Nil(err)
	s.Equal(true, GetPath(j, "deflated").(bool))
}

func (s *ResponseSuite) TestText() {
	res, err := s.c.
		Get(testHost + "/get").
		End()

	s.Nil(err)
	s.NotEmpty(res.Text())
}

func (s *ResponseSuite) TestNotOkJSON() {
	res, err := s.c.
		Get(testHost + "/status/418").
		JSON()

	s.Nil(res)
	s.Contains(err.Error(), "teapot")
}

func (s *ResponseSuite) TestNotOkText() {
	_, err := s.c.
		Get(testHost + "/post").
		Text()

	s.Equal(ErrStatusNotOk, err)
}

func (s *ResponseSuite) TestGetIndex() {
	data := []interface{}{"a", "b", "c"}

	var res1 interface{} = data
	res2 := &[]interface{}{}
	b, _ := json.Marshal(data)
	json.Unmarshal(b, res2)

	s.Equal("a", GetIndex(res1, 0).(string))
	s.Equal("a", GetIndex(res2, 0).(string))
	s.Nil(GetIndex(res1, 3))
	s.Nil(GetIndex(res2, 3))

	s.Nil(GetIndex([]string{}, 0))
	s.Nil(GetIndex(map[string]string{}, 0))
}

func (s *ResponseSuite) TestGetPath() {
	data := map[string]interface{}{
		"key1": map[string]interface{}{
			"key2": map[string]interface{}{
				"key3": 1,
				"key4": "val",
				"key5": map[int]int{1: 1},
				"key6": []interface{}{"a", "b"},
			},
		},
	}

	var res1 interface{} = data
	res2 := &map[string]interface{}{}
	b, _ := json.Marshal(data)
	json.Unmarshal(b, res2)

	s.Equal(1, GetPath(res1, "key1", "key2", "key3").(int))
	s.Equal(float64(1), GetPath(res2, "key1", "key2", "key3").(float64))

	s.Equal("val", GetPath(res1, "key1", "key2", "key4").(string))
	s.Equal("val", GetPath(res2, "key1", "key2", "key4").(string))

	s.Equal(map[int]int{1: 1}, GetPath(res1, "key1", "key2", "key5").(map[int]int))
	s.Equal(float64(1), GetPath(res2, "key1", "key2", "key5", "1").(float64))

	s.Equal("b", GetIndex(GetPath(res1, "key1", "key2", "key6"), 1).(string))
	s.Equal("b", GetIndex(GetPath(res1, "key1", "key2", "key6"), 1).(string))

	s.Nil(GetPath(res1, "key1", "key2", "key"))
	s.Nil(GetPath(res2, "key1", "key2", "key"))
	s.Nil(GetPath(res1, "key1", "key"))
	s.Nil(GetPath(res2, "key1", "key"))
	s.Nil(GetPath(res1, "key"))
	s.Nil(GetPath(res2, "key"))
	s.Nil(GetPath(res1))
	s.Nil(GetPath(res2))
	s.Nil(GetPath(map[string]string{}))
	s.Nil(GetPath([]string{}))
}

func TestResponse(t *testing.T) {
	suite.Run(t, new(ResponseSuite))
}
