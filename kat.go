// Package kat provides access to kickass.
package kat

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	noResultsMagic        = "did not match any documents"
	errExtractErrorString = "kat: did not find expected html structure"
)

type errExtractError struct {
	body string
}

func (e errExtractError) Error() string {
	return errExtractErrorString
}

// GetErrRawBody returns the raw response body from the error. When Search
// returns an error, this may be used to retrieve the raw response body if one
// is available.
func GetErrRawBody(err error) string {
	if ee, ok := err.(errExtractError); ok {
		return ee.body
	}
	return ""
}

// Result is an individual search result.
type Result struct {
	Name     string
	Magnet   string
	Verified bool
	Size     string
	Files    int
	Age      string
	Seed     int
	Leech    int
}

// Client provides access to kickass.
type Client struct {
	url       *url.URL
	transport http.RoundTripper
}

// Search performs a search query.
func (c *Client) Search(q string) ([]Result, error) {
	u, err := c.url.Parse(fmt.Sprintf("/usearch/%s/", url.QueryEscape(q)))
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{},
	}
	res, err := c.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bd, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bd))
	if err != nil {
		return nil, err
	}

	if strings.Contains(doc.Text(), noResultsMagic) {
		return nil, nil
	}

	// only use results from the first table with magnet links.
	// this prevents using "suggested search" results.
	table := doc.Find("a[href*=magnet]").First().Closest("table")

	var results []Result
	table.Find("a[href*=magnet]").Each(func(i int, s *goquery.Selection) {
		tr := s.Closest("tr")
		var r Result
		r.Magnet, _ = s.Attr("href")
		r.Name = tr.Find("td:nth-child(1) .cellMainLink").Text()
		r.Size = tr.Find("td:nth-child(2)").Text()
		r.Files, _ = strconv.Atoi(tr.Find("td:nth-child(3)").Text())
		r.Age = tr.Find("td:nth-child(4)").Text()
		r.Seed, _ = strconv.Atoi(tr.Find("td:nth-child(4)").Text())
		r.Leech, _ = strconv.Atoi(tr.Find("td:nth-child(5)").Text())
		results = append(results, r)
	})

	if len(results) == 0 {
		return nil, errExtractError{body: string(bd)}
	}

	return results, nil
}

// ClientOption allows configuring various aspects of the Client.
type ClientOption func(*Client) error

// ClientTransport configures the Transport for the Client. If not specified
// http.DefaultTransport is used.
func ClientTransport(t http.RoundTripper) ClientOption {
	return func(c *Client) error {
		c.transport = t
		return nil
	}
}

// ClientURL configures the base Client URL. All requests are made relative
// to this URL.
func ClientURL(u *url.URL) ClientOption {
	return func(c *Client) error {
		c.url = u
		return nil
	}
}

// ClientRawURL configures the base Client URL. All API requests are made
// relative to this URL.
func ClientRawURL(u string) ClientOption {
	return func(c *Client) error {
		var err error
		c.url, err = url.Parse(u)
		return err
	}
}

// NewClient creates a new client with the given options.
func NewClient(options ...ClientOption) (*Client, error) {
	var c Client
	for _, o := range options {
		if err := o(&c); err != nil {
			return nil, err
		}
	}
	if c.url == nil {
		c.url = &url.URL{
			Scheme: "https",
			Host:   "kickass.to",
			Path:   "/",
		}
	}
	if c.transport == nil {
		c.transport = http.DefaultTransport
	}
	return &c, nil
}
