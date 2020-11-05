package req

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var _ Requester = &request{}

// Requester HTTP request interface
type Requester interface {
	Head(ctx context.Context, urlStr string, queryParam url.Values, opts ...RequestOption) (Responser, error)
	Get(ctx context.Context, urlStr string, queryParam url.Values, opts ...RequestOption) (Responser, error)
	Delete(ctx context.Context, urlStr string, queryParam url.Values, opts ...RequestOption) (Responser, error)
	Patch(ctx context.Context, urlStr string, queryParam url.Values, opts ...RequestOption) (Responser, error)
	Post(ctx context.Context, urlStr string, body io.Reader, opts ...RequestOption) (Responser, error)
	PostJSON(ctx context.Context, urlStr string, body interface{}, opts ...RequestOption) (Responser, error)
	PostForm(ctx context.Context, urlStr string, body url.Values, opts ...RequestOption) (Responser, error)
	Put(ctx context.Context, urlStr string, body io.Reader, opts ...RequestOption) (Responser, error)
	PutJSON(ctx context.Context, urlStr string, body interface{}, opts ...RequestOption) (Responser, error)
	PutForm(ctx context.Context, urlStr string, body url.Values, opts ...RequestOption) (Responser, error)
	Do(ctx context.Context, urlStr, method string, body io.Reader, opts ...RequestOption) (Responser, error)
}

// RequestURL get request url
func RequestURL(base, router string) string {
	var buf bytes.Buffer
	if l := len(base); l > 0 {
		if base[l-1] == '/' {
			base = base[:l-1]
		}
		buf.WriteString(base)

		if rl := len(router); rl > 0 {
			if router[0] != '/' {
				buf.WriteByte('/')
			}
		}
	}
	buf.WriteString(router)
	return buf.String()
}

// New create a request instance
func New(opt ...Option) Requester {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	req := &request{
		opts: opts,
		cli: &http.Client{
			Transport:     opts.transport,
			CheckRedirect: opts.checkRedirect,
			Jar:           opts.cookieJar,
			Timeout:       opts.timeout,
		},
	}

	return req
}

type request struct {
	opts options
	cli  *http.Client
}

func (r *request) parseQueryParam(urlStr string, param url.Values) string {
	if param != nil {
		c := '?'
		if strings.IndexByte(urlStr, '?') != -1 {
			c = '&'
		}
		urlStr = fmt.Sprintf("%s%c%s", urlStr, c, param.Encode())
	}
	return urlStr
}

func (r *request) fillRequest(req *http.Request, opts ...RequestOption) (*http.Request, error) {
	if req.Header == nil {
		req.Header = make(http.Header)
	}

	for k := range r.opts.header {
		req.Header.Set(k, r.opts.header.Get(k))
	}

	ro := &requestOptions{
		request: req,
	}
	for _, opt := range opts {
		opt(ro)
	}

	if fn := ro.handle; fn != nil {
		return fn(req)
	}

	return req, nil
}

func (r *request) doForm(ctx context.Context, urlStr, method string, body url.Values, opts ...RequestOption) (Responser, error) {
	var s string
	if body != nil {
		s = body.Encode()
	}

	var ro []RequestOption
	ro = append(ro, SetContentType(MIMEApplicationForm))
	if len(opts) > 0 {
		ro = append(ro, opts...)
	}

	return r.Do(ctx, urlStr, method, strings.NewReader(s), ro...)
}

func (r *request) doJSON(ctx context.Context, urlStr, method string, body interface{}, opts ...RequestOption) (Responser, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return nil, err
	}

	var ro []RequestOption
	ro = append(ro, SetContentType(MIMEApplicationJSONCharsetUTF8))
	if len(opts) > 0 {
		ro = append(ro, opts...)
	}
	return r.Do(ctx, urlStr, method, buf, ro...)
}

func (r *request) httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	return f(r.cli.Do(req))
}

func (r *request) Head(ctx context.Context, urlStr string, queryParam url.Values, opts ...RequestOption) (Responser, error) {
	return r.Do(ctx, r.parseQueryParam(urlStr, queryParam), http.MethodHead, nil, opts...)
}

func (r *request) Get(ctx context.Context, urlStr string, queryParam url.Values, opts ...RequestOption) (Responser, error) {
	return r.Do(ctx, r.parseQueryParam(urlStr, queryParam), http.MethodGet, nil, opts...)
}

func (r *request) Delete(ctx context.Context, urlStr string, queryParam url.Values, opts ...RequestOption) (Responser, error) {
	return r.Do(ctx, r.parseQueryParam(urlStr, queryParam), http.MethodDelete, nil, opts...)
}

func (r *request) Patch(ctx context.Context, urlStr string, queryParam url.Values, opts ...RequestOption) (Responser, error) {
	return r.Do(ctx, r.parseQueryParam(urlStr, queryParam), http.MethodPatch, nil, opts...)
}

func (r *request) Post(ctx context.Context, urlStr string, body io.Reader, opts ...RequestOption) (Responser, error) {
	return r.Do(ctx, urlStr, http.MethodPost, body, opts...)
}

func (r *request) PostJSON(ctx context.Context, urlStr string, body interface{}, opts ...RequestOption) (Responser, error) {
	return r.doJSON(ctx, urlStr, http.MethodPost, body, opts...)
}

func (r *request) PostForm(ctx context.Context, urlStr string, body url.Values, opts ...RequestOption) (Responser, error) {
	return r.doForm(ctx, urlStr, http.MethodPost, body, opts...)
}

func (r *request) Put(ctx context.Context, urlStr string, body io.Reader, opts ...RequestOption) (Responser, error) {
	return r.Do(ctx, urlStr, http.MethodPut, body, opts...)
}

func (r *request) PutJSON(ctx context.Context, urlStr string, body interface{}, opts ...RequestOption) (Responser, error) {
	return r.doJSON(ctx, urlStr, http.MethodPut, body, opts...)
}

func (r *request) PutForm(ctx context.Context, urlStr string, body url.Values, opts ...RequestOption) (Responser, error) {
	return r.doForm(ctx, urlStr, http.MethodPut, body, opts...)
}

func (r *request) Do(ctx context.Context, urlStr, method string, body io.Reader, opts ...RequestOption) (Responser, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	url := RequestURL(r.opts.baseURL, urlStr)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req, err = r.fillRequest(req, opts...)
	if err != nil {
		return nil, err
	}

	var resp Responser
	err = r.httpDo(ctx, req, func(res *http.Response, err error) error {
		if err != nil {
			return err
		}
		resp = newResponse(res)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
