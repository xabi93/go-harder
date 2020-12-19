package aimharder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

const (
	defaultBaseURL   = "https://aimharder.com/aplic/apiExternal/"
	defaultUserAgent = "go-aimharder"
)

type Logger interface {
	Println(v ...interface{})
}

// Client manages comunication with Aimharder api
type Client struct {
	client *http.Client
	debug  bool
	logger Logger

	baseUrl   *url.URL
	userAgent string

	user *User

	authToken string
	// Services to call different endpoints of Aimharder API
	Users    *UsersService
	Bookings *BookingsService
}

type service struct {
	client *Client
}

func OptionBaseURL(url *url.URL) ClientOption {
	return func(c *Client) {
		c.baseUrl = url
	}
}

func OptionHttpClient(httpCli *http.Client) ClientOption {
	return func(c *Client) {
		c.client = httpCli
	}
}

func OptionDebug(d bool) ClientOption {
	return func(c *Client) {
		c.debug = d
	}
}

func OptionLogger(l Logger) ClientOption {
	return func(c *Client) {
		c.logger = l
	}
}

type ClientOption func(*Client)

// New creates a new aimharder client with auth token
func New(token string, opts ...ClientOption) (*Client, error) {
	baseUrl, _ := url.Parse(defaultBaseURL)

	client := Client{
		client:    &http.Client{},
		baseUrl:   baseUrl,
		userAgent: defaultUserAgent,
		authToken: token,
		logger:    log.New(os.Stderr, "go-harder", 0),
	}
	for _, opt := range opts {
		opt(&client)
	}

	if !strings.HasSuffix(client.baseUrl.String(), "/") {
		return nil, fmt.Errorf("base url must have a trailing slash, %q", client.baseUrl)
	}

	s := service{client: &client}

	// init services
	client.Users = (*UsersService)(&s)
	client.Bookings = (*BookingsService)(&s)

	return &client, nil
}

// Login logs in the application given a user email and password and returns the client
func Login(ctx context.Context, mail, password string, opts ...ClientOption) (*Client, error) {
	c, err := New("", opts...)
	if err != nil {
		return nil, err
	}

	if err := c.login(ctx, mail, password); err != nil {
		return nil, err
	}

	return c, nil
}

func (c Client) AuthToken() string {
	return c.authToken
}

func (c *Client) newRequest(ctx context.Context, method string, endpoint string, body interface{}) (*http.Request, error) {
	url, err := c.baseUrl.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint error: %w", err)
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		e := json.NewEncoder(buf)
		if err := e.Encode(body); err != nil {
			return nil, fmt.Errorf("encode request body error: %w", err)
		}
	}

	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("http NewRequest error: %w", err)
	}
	req = req.WithContext(ctx)

	if c.authToken != "" {
		q := req.URL.Query()
		q.Set("token", c.authToken)
		req.URL.RawQuery = q.Encode()
	}

	if body != nil {
		req.Header.Set("content-type", "application/json")
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

type apiError struct {
	Logout *int    `json:"logout,omitempty"`
	Err    *string `json:"error,omitempty"`
}

func (err apiError) Error() error {
	if err.Err == nil && err.Logout == nil {
		return nil
	}

	if err.Logout != nil {
		return LogoutError
	}

	switch *err.Err {
	case apiInvalidMailPassError:
		return InvalidMailPassLoginError
	}

	return unknownError(*err.Err)

}

var _ json.Unmarshaler = &apiBodyDecoder{}

type apiBodyDecoder struct {
	success interface{}

	apiError apiError
}

func (d *apiBodyDecoder) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	if err := json.Unmarshal(data, &d.apiError); err != nil {
		return err
	}

	if d.apiError.Error() != nil {
		return nil
	}

	return json.Unmarshal(data, d.success)
}

func (d apiBodyDecoder) Error() error {
	return d.apiError.Error()
}

func (c *Client) do(ctx context.Context, req *http.Request, decoder *apiBodyDecoder) (*http.Response, error) {
	if c.debug {
		text, _ := httputil.DumpRequest(req, true)
		c.logger.Println(string(text))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// Maybe the error is because the ctx is cancelled so we prioritize that error
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, fmt.Errorf("http.Do error: %w", err)
	}
	defer resp.Body.Close()

	if c.debug {
		text, _ := httputil.DumpResponse(resp, true)
		c.logger.Println(string(text))
	}

	// The api always returns 200 status code, we can assume if it returns 404 is because the endpoint does not exists
	if resp.StatusCode == http.StatusNotFound {
		return resp, fmt.Errorf("endpoint %s: %w", req.URL, EndpointNotExistsError)
	}

	if resp.Body != nil {
		err := json.NewDecoder(resp.Body).Decode(&decoder)
		if err != nil {
			return nil, fmt.Errorf("decode response body error: %w", err)
		}
	}

	return resp, nil
}

// Login returns and saves auth token
func (c *Client) login(ctx context.Context, email, password string) error {
	req, err := c.newRequest(ctx, http.MethodGet, "login", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Set("mail", email)
	q.Set("pw", password)
	req.URL.RawQuery = q.Encode()

	var success struct {
		Cookie string `json:"cookie"`
	}
	result := apiBodyDecoder{
		success: &success,
	}

	_, err = c.do(ctx, req, &result)
	if err != nil {
		return err
	}

	if err := result.Error(); err != nil {
		return err
	}

	if success.Cookie == "" {
		return MissingAuthTokenError
	}

	c.authToken = success.Cookie

	return nil
}
