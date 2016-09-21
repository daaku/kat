package kat

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/facebookgo/ensure"
)

func TestErrExtractErrorString(t *testing.T) {
	ensure.DeepEqual(t, (errExtractError{}).Error(), errExtractErrorString)
	ensure.DeepEqual(t, (errExtractError{body: "foo"}).Error(), errExtractErrorString)
}

func TestGetErrRawBody(t *testing.T) {
	const body = "body"
	ensure.DeepEqual(t, GetErrRawBody(errExtractError{body: body}), body)
}

func TestNoGetErrRawBody(t *testing.T) {
	ensure.DeepEqual(t, GetErrRawBody(errors.New("foo")), "")
}

type fTransport func(*http.Request) (*http.Response, error)

func (f fTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestNoResults(t *testing.T) {
	c, err := NewClient(
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader("<html>")),
			}, nil
		})),
	)
	ensure.Nil(t, err)
	res, err := c.Search("unimportant")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(res), 0)
}

func TestTransportError(t *testing.T) {
	givenErr := errors.New("")
	c, err := NewClient(
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return nil, givenErr
		})),
	)
	ensure.Nil(t, err)
	res, err := c.Search("unimportant")
	ensure.DeepEqual(t, err, givenErr)
	ensure.DeepEqual(t, len(res), 0)
}

func TestBodyReadError(t *testing.T) {
	givenErr := errors.New("")
	r, w := io.Pipe()
	w.CloseWithError(givenErr)
	c, err := NewClient(
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{Body: ioutil.NopCloser(r)}, nil
		})),
	)
	ensure.Nil(t, err)
	res, err := c.Search("unimportant")
	ensure.DeepEqual(t, err, givenErr)
	ensure.DeepEqual(t, len(res), 0)
}

func TestNormalResults(t *testing.T) {
	c, err := NewClient(
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`
				<body>
					<table>
						<tr>
							<td>
								<a href="magnet:a">l</a>
							</td>
						</tr>
					</table>
				</body>
			`)),
			}, nil
		})),
	)
	ensure.Nil(t, err)
	res, err := c.Search("unimportant")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, res, []Result{{Magnet: "magnet:a"}})
}

func TestIgnoreSuggestedResults(t *testing.T) {
	c, err := NewClient(
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`
				<body>
					<table>
						<tr>
							<td>
								<a href="magnet:a">l</a>
							</td>
						</tr>
					</table>
					<table>
						<tr>
							<td>
								<a href="magnet:b">l</a>
							</td>
						</tr>
					</table>
				</body>
			`)),
			}, nil
		})),
	)
	ensure.Nil(t, err)
	res, err := c.Search("unimportant")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, res, []Result{{Magnet: "magnet:a"}})
}

func TestNewClientError(t *testing.T) {
	givenErr := errors.New("")
	c, err := NewClient(func(*Client) error { return givenErr })
	ensure.True(t, c == nil)
	ensure.DeepEqual(t, err, givenErr)
}

func TestNewClientURL(t *testing.T) {
	u := &url.URL{}
	c, err := NewClient(ClientURL(u))
	ensure.Nil(t, err)
	ensure.DeepEqual(t, c.url, u)
}

func TestNewClientRawURL(t *testing.T) {
	c, err := NewClient(ClientRawURL("http://foo.com"))
	ensure.Nil(t, err)
	ensure.DeepEqual(t, c.url, &url.URL{
		Scheme: "http",
		Host:   "foo.com",
	})
}

func TestEmptyQuerySearch(t *testing.T) {
	c, err := NewClient(ClientRawURL("http://foo.com"))
	ensure.Nil(t, err)
	res, err := c.Search("")
	ensure.DeepEqual(t, err, errEmptyQuery)
	ensure.DeepEqual(t, len(res), 0)
}
