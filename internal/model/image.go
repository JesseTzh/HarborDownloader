package model

type RegistryConfig struct {
	Registry string `json:"registry"`
	Username string `json:"username"`
	Password string `json:"password"`
	Insecure bool   `json:"insecure"`
}

type ImageReference struct {
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Digest     string `json:"digest,omitempty"`
	Raw        string `json:"raw"`
}

func (r ImageReference) String() string {
	if r.Raw != "" {
		return r.Raw
	}
	if r.Digest != "" {
		return r.Registry + "/" + r.Repository + "@" + r.Digest
	}
	tag := r.Tag
	if tag == "" {
		tag = "latest"
	}
	return r.Registry + "/" + r.Repository + ":" + tag
}

type Platform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant,omitempty"`
}

func (p Platform) String() string {
	if p.OS == "" && p.Architecture == "" {
		return ""
	}
	s := p.OS + "/" + p.Architecture
	if p.Variant != "" {
		s += "/" + p.Variant
	}
	return s
}

func DefaultPlatform() Platform {
	return Platform{OS: "linux", Architecture: "amd64"}
}

func SupportedPlatforms() []Platform {
	return []Platform{
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
	}
}
