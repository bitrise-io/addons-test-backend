package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Interface ...
type Interface interface {
	Get(*http.Request, string) (interface{}, error)
	Set(string) error
	Clear()
	Save() error
}

// Client ...
type Client struct {
	session *sessions.Session
	req     *http.Request
	resp    http.ResponseWriter
}

// NewClient ...
func NewClient(session *sessions.Session, r *http.Request, w http.ResponseWriter) Client {
	return Client{session: session, req: r, resp: w}
}

// Get ...
func (c *Client) Get(key interface{}) interface{} {
	return c.session.Values[key]
}

// Set ...
func (c *Client) Set(key, value interface{}) {
	c.session.Values[key] = value
}

// Clear ...
func (c *Client) Clear(r *http.Request) {
	for _, value := range c.session.Values {
		delete(c.session.Values, value)
	}
}

// Save ...
func (c *Client) Save() error {
	return c.session.Save(c.req, c.resp)
}
