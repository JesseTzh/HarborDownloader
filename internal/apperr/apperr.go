package apperr

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

const (
	CodeUnreachable  = "unreachable"
	CodeUnauthorized = "unauthorized"
	CodeForbidden    = "forbidden"
	CodeNotFound     = "not_found"
	CodeTLS          = "tls"
	CodeDiskFull     = "disk_full"
	CodeCancelled    = "cancelled"
	CodeInvalid      = "invalid"
	CodeUnknown      = "unknown"
)

type Error struct {
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func New(code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

func Classify(err error) *Error {
	if err == nil {
		return nil
	}
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	if errors.Is(err, context.Canceled) {
		return New(CodeCancelled, "任务已取消。", err)
	}

	msg := err.Error()
	lower := strings.ToLower(msg)
	status := httpStatus(err)

	if isTLS(lower) {
		return New(CodeTLS, "TLS 证书验证失败。\n\n请确认：\n- Harbor CA 是否已加入系统信任\n- 地址是否正确", err)
	}
	if isDiskFull(err, lower) {
		return New(CodeDiskFull, "磁盘空间不足。", err)
	}
	if status == 401 || isUnauthorized(lower) {
		return New(CodeUnauthorized, "认证失败。\n\n用户名、密码或 Robot Account Secret 不正确。", err)
	}
	if status == 403 || isForbidden(lower) {
		return New(CodeForbidden, "没有权限拉取该 Repository。", err)
	}
	if status == 404 || isNotFound(lower) {
		return New(CodeNotFound, "镜像不存在。\n\n请检查 Repository 和 Tag。", err)
	}
	if isUnreachable(err, lower) {
		return New(CodeUnreachable, "无法连接 Harbor。\n\n请检查：\n1. Harbor 地址\n2. VPN\n3. DNS\n4. 网络连接", err)
	}

	return New(CodeUnknown, "操作失败：\n"+sanitize(msg), err)
}

func sanitize(msg string) string {
	for _, key := range []string{"password", "token", "secret", "authorization"} {
		if strings.Contains(strings.ToLower(msg), key) {
			return "内部错误，详情已写入日志。"
		}
	}
	if len(msg) > 500 {
		return msg[:500] + "…"
	}
	return msg
}

func httpStatus(err error) int {
	var te *transport.Error
	if errors.As(err, &te) {
		return te.StatusCode
	}
	return 0
}

func isTLS(lower string) bool {
	return strings.Contains(lower, "tls:") ||
		strings.Contains(lower, "x509:") ||
		strings.Contains(lower, "certificate verify") ||
		strings.Contains(lower, "certificate is not trusted") ||
		strings.Contains(lower, "unknown authority")
}

func isDiskFull(err error, lower string) bool {
	if errors.Is(err, syscall.ENOSPC) {
		return true
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) && errors.Is(pathErr.Err, syscall.ENOSPC) {
		return true
	}
	return strings.Contains(lower, "no space left")
}

func isUnauthorized(lower string) bool {
	return strings.Contains(lower, "unauthorized") ||
		strings.Contains(lower, "authentication required")
}

func isForbidden(lower string) bool {
	return strings.Contains(lower, "forbidden") ||
		strings.Contains(lower, "access denied") ||
		strings.Contains(lower, "denied")
}

func isNotFound(lower string) bool {
	return strings.Contains(lower, "manifest unknown") ||
		strings.Contains(lower, "name unknown") ||
		strings.Contains(lower, "not found")
}

func isUnreachable(err error, lower string) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	if errors.Is(err, io.EOF) && (strings.Contains(lower, "connect") || strings.Contains(lower, "dial")) {
		return true
	}
	return strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "no such host") ||
		strings.Contains(lower, "i/o timeout") ||
		strings.Contains(lower, "network is unreachable") ||
		strings.Contains(lower, "connection reset") ||
		strings.Contains(lower, "dial tcp")
}
