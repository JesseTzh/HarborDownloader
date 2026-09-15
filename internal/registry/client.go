package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"HarborDownloader/internal/apperr"
	"HarborDownloader/internal/model"
)

type Client struct {
	auth      authn.Authenticator
	transport http.RoundTripper
	insecure  bool
	http      bool
}

func NewClient(cfg model.RegistryConfig) *Client {
	n, _ := NormalizeRegistry(cfg.Registry, cfg.Insecure)
	return &Client{
		auth:      Authenticator(cfg.Username, cfg.Password),
		transport: HTTPTransport(cfg.Insecure),
		insecure:  cfg.Insecure,
		http:      n.HTTP,
	}
}

func (c *Client) remoteOptions(ctx context.Context, platform model.Platform) []remote.Option {
	opts := []remote.Option{
		remote.WithAuth(c.auth),
		remote.WithContext(ctx),
		remote.WithTransport(c.transport),
	}
	if platform.OS != "" && platform.Architecture != "" {
		opts = append(opts, remote.WithPlatform(ggcrPlatform(platform)))
	}
	return opts
}

func (c *Client) nameOptions() []name.Option {
	opts := []name.Option{name.WeakValidation}
	if c.insecure || c.http {
		opts = append(opts, name.Insecure)
	}
	return opts
}

func (c *Client) registry(host string) (name.Registry, error) {
	return name.NewRegistry(host, c.nameOptions()...)
}

func (c *Client) Test(ctx context.Context, cfg model.RegistryConfig) model.TestRegistryResult {
	out := model.TestRegistryResult{}
	n, err := NormalizeRegistry(cfg.Registry, cfg.Insecure)
	if err != nil {
		out.Error = apperr.Classify(err).Message
		out.Message = out.Error
		return out
	}

	reg, err := name.NewRegistry(n.Host, c.nameOptions()...)
	if err != nil {
		out.Error = apperr.Classify(apperr.New(apperr.CodeInvalid, "Harbor 地址无效。", err)).Message
		out.Message = out.Error
		return out
	}

	pingCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	_, err = transport.Ping(pingCtx, reg, c.transport)
	if err != nil {
		classified := apperr.Classify(err)
		out.Error = classified.Message
		out.Message = classified.Message
		if classified.Code == apperr.CodeUnauthorized {
			out.RegistryReachable = true
		}
		return out
	}
	out.RegistryReachable = true

	tr, err := transport.NewWithContext(pingCtx, reg, c.auth, c.transport, []string{})
	if err != nil {
		classified := apperr.Classify(err)
		out.Error = classified.Message
		out.Message = classified.Message
		return out
	}

	client := &http.Client{Transport: tr, Timeout: 20 * time.Second}
	u := fmt.Sprintf("%s://%s/v2/", reg.Scheme(), reg.Name())
	req, err := http.NewRequestWithContext(pingCtx, http.MethodGet, u, nil)
	if err != nil {
		out.Error = apperr.Classify(err).Message
		out.Message = out.Error
		return out
	}
	resp, err := client.Do(req)
	if err != nil {
		classified := apperr.Classify(err)
		out.Error = classified.Message
		out.Message = classified.Message
		return out
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		out.Error = apperr.Classify(apperr.New(apperr.CodeUnauthorized, "unauthorized", nil)).Message
		out.Message = out.Error
		return out
	}
	if resp.StatusCode >= 400 {
		out.Error = apperr.Classify(fmt.Errorf("registry ping status %d", resp.StatusCode)).Message
		out.Message = out.Error
		return out
	}
	out.AuthSuccess = true

	out.HarborReachable = c.pingHarborAPI(pingCtx, n, cfg.Username, cfg.Password)

	if out.AuthSuccess && out.RegistryReachable {
		out.OK = true
		var b strings.Builder
		b.WriteString("✓ Registry 可连接\n")
		b.WriteString("✓ 认证成功\n")
		if out.HarborReachable {
			b.WriteString("✓ Harbor 可访问")
		} else {
			b.WriteString("• 未检测到 Harbor API（Registry 仍可拉取镜像）")
		}
		out.Message = b.String()
	}
	return out
}

func (c *Client) pingHarborAPI(ctx context.Context, n NormalizedRegistry, username, password string) bool {
	client := &http.Client{Transport: c.transport, Timeout: 10 * time.Second}
	for _, path := range []string{"/api/v2.0/health", "/api/v2.0/systeminfo"} {
		u := fmt.Sprintf("%s://%s%s", n.Scheme, n.Host, path)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			continue
		}
		if username != "" || password != "" {
			req.SetBasicAuth(username, password)
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if path == "/api/v2.0/health" {
				var payload map[string]any
				_ = json.Unmarshal(body, &payload)
			}
			return true
		}
	}
	return false
}
