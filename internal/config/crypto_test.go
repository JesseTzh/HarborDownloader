package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestEncryptExportRoundTrip(t *testing.T) {
	plain := []byte(`{"registry":"harbor.local","password":"s3cret"}`)
	enc, err := EncryptExport(plain)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(enc, []byte("s3cret")) {
		t.Fatal("password leaked in ciphertext envelope")
	}
	var env exportEnvelope
	if err := json.Unmarshal(enc, &env); err != nil {
		t.Fatal(err)
	}
	if env.Kind != exportKind || env.V != exportVersion || env.N == "" || env.C == "" {
		t.Fatalf("%+v", env)
	}
	got, err := openExport(enc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("got %s", got)
	}
}

func TestOpenExportLeavesPlainJSON(t *testing.T) {
	plain := []byte(`{"registry":"harbor.local","password":"s3cret"}`)
	got, err := openExport(plain)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("got %s", got)
	}
}

func TestOpenExportRejectsTamperedCiphertext(t *testing.T) {
	enc, err := EncryptExport([]byte(`{"registry":"harbor.local"}`))
	if err != nil {
		t.Fatal(err)
	}
	var env exportEnvelope
	if err := json.Unmarshal(enc, &env); err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(env.C)
	if err != nil || len(raw) == 0 {
		t.Fatal(err)
	}
	raw[0] ^= 0xff
	env.C = base64.StdEncoding.EncodeToString(raw)
	tampered, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := openExport(tampered); err == nil {
		t.Fatal("expected decrypt error")
	}
}

func TestOpenExportRejectsUnsupportedVersion(t *testing.T) {
	b, err := json.Marshal(exportEnvelope{Kind: exportKind, V: 99, N: "YQ==", C: "YQ=="})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := openExport(b); err == nil {
		t.Fatal("expected version error")
	}
}
