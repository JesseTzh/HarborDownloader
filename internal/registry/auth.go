package registry

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

func Authenticator(username, password string) authn.Authenticator {
	if username == "" && password == "" {
		return authn.Anonymous
	}
	return &authn.Basic{Username: username, Password: password}
}

func HTTPTransport(insecure bool) http.RoundTripper {
	base, ok := http.DefaultTransport.(*http.Transport)
	var t *http.Transport
	if ok {
		t = base.Clone()
	} else {
		t = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}
	if insecure {
		cfg := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-opt-in for self-signed Harbor
		t.TLSClientConfig = cfg
	}
	return transport.NewRetry(t,
		transport.WithRetryBackoff(transport.Backoff{
			Duration: 1 * time.Second,
			Factor:   2.0,
			Jitter:   0.1,
			Steps:    3,
		}),
		transport.WithRetryStatusCodes(429, 500, 502, 503, 504),
	)
}
