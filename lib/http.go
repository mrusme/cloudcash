package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"golang.org/x/oauth2"
)

const UserAgent = "xn--gckvb8fzb.com/cloudcash"

type statusCodeKey struct{}

type statusCodeTransport struct {
	base http.RoundTripper
}

func NewTransport() http.RoundTripper {
	return statusCodeTransport{base: http.DefaultTransport}
}

func NewHTTPClient() *http.Client {
	return &http.Client{Transport: NewTransport()}
}

func NewBearerHTTPClient(token string) *http.Client {
	return &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
			Base:   NewTransport(),
		},
	}
}

func (t statusCodeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if code, ok := req.Context().Value(statusCodeKey{}).(*atomic.Int32); ok {
		code.Store(int32(resp.StatusCode))
	}

	return resp, nil
}

func WithStatusCodeRecorder(ctx context.Context) (context.Context, func() int) {
	code := new(atomic.Int32)
	return context.WithValue(ctx, statusCodeKey{}, code), func() int {
		return int(code.Load())
	}
}

func GetJSON(
	ctx context.Context,
	client *http.Client,
	url string,
	headers map[string]string,
	v any,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			return fmt.Errorf(
				"%s returned %s, retry after %s",
				url,
				resp.Status,
				retryAfter,
			)
		}
		return fmt.Errorf("%s returned %s", url, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}
