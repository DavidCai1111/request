package request

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

const testHost = "http://httpbin.org"

type RequestSuite struct {
	suite.Suite

	c *Client
}

func (s *RequestSuite) SetupTest() {
	s.c = New(nil)
}

func (s *RequestSuite) TestGet() {
	res, err := s.c.Get(testHost).End()

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
	s.Equal(body["k1"], j.GetPath("json", "k1").MustString())
	s.Equal(body["k2"], j.GetPath("json", "k2").MustString())
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
	s.Equal(body["k1"], j.GetPath("json", "k1").MustString())
	s.Equal(body["k2"], j.GetPath("json", "k2").MustString())
}

func (s *RequestSuite) TestDelete() {
	res, err := s.c.
		Delete(testHost + "/delete?k1=v1&k2=v2").
		End()

	s.Nil(err)
	s.True(res.OK())
}

func TestRequest(t *testing.T) {
	suite.Run(t, new(RequestSuite))
}
