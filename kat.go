// Package kat provides access to kickass.
package kat

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	errExtractErrorString = "kat: did not find expected html structure"
)

var errEmptyQuery = errors.New("kat: invalid empty query search")

// Result is an individual search result.
type Result struct {
	Name     string
	Magnet   string
	Verified bool
	Size     string
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
	if q == "" {
		return nil, errEmptyQuery
	}

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
	// TODO: error on non 200 responses?

	bd, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bd))
	if err != nil {
		return nil, err
	}

	// only use results from the first table with magnet links.
	// this prevents using "suggested search" results.
	table := doc.Find("a[href*=magnet]").First().Closest("table")

	var results []Result
	table.Find("a[href*=magnet]").Each(func(i int, s *goquery.Selection) {
		tr := s.Closest("tr")
		var r Result
		r.Magnet, _ = s.Attr("href")
		r.Name = strings.TrimSpace(tr.Find("td:nth-child(1) .cellMainLink").Text())
		r.Size = strings.TrimSpace(tr.Find("td:nth-child(2)").Text())
		r.Age = strings.TrimSpace(tr.Find("td:nth-child(3)").Text())
		r.Seed, _ = strconv.Atoi(tr.Find("td:nth-child(4)").Text())
		r.Leech, _ = strconv.Atoi(tr.Find("td:nth-child(5)").Text())
		results = append(results, r)
	})

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
