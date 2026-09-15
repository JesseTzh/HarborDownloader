package model

import "testing"

func TestParseTargetTag(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantRepo string
		wantTag  string
		wantRaw  string
		wantErr  bool
	}{
		{name: "empty", raw: "  "},
		{
			name:     "short name",
			raw:      "app-gateway:1.0.0",
			wantRepo: "library/app-gateway",
			wantTag:  "1.0.0",
			wantRaw:  "app-gateway:1.0.0",
		},
		{
			name:     "namespaced",
			raw:      "company/app:prod",
			wantRepo: "company/app",
			wantTag:  "prod",
			wantRaw:  "company/app:prod",
		},
		{
			name:     "with registry",
			raw:      "harbor.company.local/project/app:1.0.0",
			wantRepo: "project/app",
			wantTag:  "1.0.0",
			wantRaw:  "harbor.company.local/project/app:1.0.0",
		},
		{
			name:    "digest rejected",
			raw:     "app@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: true,
		},
		{
			name:    "path traversal",
			raw:     "../evil:1",
			wantErr: true,
		},
		{
			name:    "invalid",
			raw:     "UPPER CASE:tag",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTargetTag(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantRaw == "" {
				if got.Raw != "" || got.Repository != "" {
					t.Fatalf("empty target should parse as empty, got %+v", got)
				}
				return
			}
			if got.Repository != tt.wantRepo {
				t.Errorf("repository=%s want %s", got.Repository, tt.wantRepo)
			}
			if got.Tag != tt.wantTag {
				t.Errorf("tag=%s want %s", got.Tag, tt.wantTag)
			}
			if got.Raw != tt.wantRaw {
				t.Errorf("raw=%s want %s", got.Raw, tt.wantRaw)
			}
		})
	}
}
