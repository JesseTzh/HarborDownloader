package apperr

import (
	"context"
	"errors"
	"syscall"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{
		{context.Canceled, CodeCancelled},
		{errors.New("x509: certificate signed by unknown authority"), CodeTLS},
		{errors.New("unauthorized: authentication required"), CodeUnauthorized},
		{errors.New("denied: requested access to the resource is denied"), CodeForbidden},
		{errors.New("manifest unknown: tag not found"), CodeNotFound},
		{errors.New("dial tcp: lookup harbor.company.local: no such host"), CodeUnreachable},
		{syscall.ENOSPC, CodeDiskFull},
	}
	for _, c := range cases {
		got := Classify(c.err)
		if got.Code != c.code {
			t.Errorf("Classify(%v)=%s want %s msg=%s", c.err, got.Code, c.code, got.Message)
		}
		if got.Message == "" || got.Message == "download failed" {
			t.Errorf("message too generic: %q", got.Message)
		}
	}
}

func TestClassifyDoesNotLeakSecrets(t *testing.T) {
	got := Classify(errors.New("failed with password=hunter2"))
	if got.Code != CodeUnknown {
		t.Fatalf("code %s", got.Code)
	}
	if got.Message == "failed with password=hunter2" {
		t.Fatal("leaked password")
	}
}

func TestConfigInvalid(t *testing.T) {
	err := New(CodeInvalid, "Harbor 地址不能为空。", nil)
	got := Classify(err)
	if got.Code != CodeInvalid {
		t.Fatal(got.Code)
	}
}
