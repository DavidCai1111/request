package request

import (
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
	s.c = New(nil)
}

func (s *ResponseSuite) TestGet() {
	res, err := s.c.Get(testHost).End()

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
	s.Equal("OK", res.Reason())
}

func (s *ResponseSuite) TestURL() {
	res, err := s.c.Get(testHost).End()

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)

	u, err := res.URL()

	s.Nil(err)
	s.Equal(testHost, u.String())
}

func (s *ResponseSuite) TestGzip() {
	j, err := s.c.
		Get(testHost+"/gzip").
		Set(headers.AcceptEncoding, "gzip, deflate").
		JSON()

	s.Nil(err)
	s.Equal(true, j.GetPath("gzipped").MustBool())
}

func (s *ResponseSuite) TestDeflate() {
	j, err := s.c.
		Get(testHost + "/deflate").
		JSON()

	s.Nil(err)
	s.Equal(true, j.GetPath("deflated").MustBool())
}

func (s *ResponseSuite) TestText() {
	res, err := s.c.
		Get(testHost + "/get").
		End()

	s.Nil(err)
	s.NotEmpty(res.Text())
}

func (s *ResponseSuite) TestNotOkJSON() {
	_, err := s.c.
		Get(testHost + "/post").
		JSON()

	s.Equal(ErrStatusNotOk, err)
}

func (s *ResponseSuite) TestNotOkText() {
	_, err := s.c.
		Get(testHost + "/post").
		Text()

	s.Equal(ErrStatusNotOk, err)
}

func TestResponse(t *testing.T) {
	suite.Run(t, new(ResponseSuite))
}
