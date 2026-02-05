package main

import (
	"testing"
)

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid base64",
			input:   "SGVsbG8gV29ybGQ=",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "base64 with newlines",
			input:   "SGVsbG8g\nV29ybGQ=",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "base64 with spaces",
			input:   "SGVsbG8g V29ybGQ=",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "base64 with tabs and newlines",
			input:   "SGVsbG8g\tV29ybGQ=\n",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "UTF-8 BOM prefix",
			input:   "\xEF\xBB\xBFSGVsbG8gV29ybGQ=",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "data URL prefix",
			input:   "data:application/octet-stream;base64,SGVsbG8gV29ybGQ=",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "base64, prefix",
			input:   "base64,SGVsbG8gV29ybGQ=",
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "URL-safe encoding",
			input:   "SGVsbG8gV29ybGQ", // no padding, URL-safe uses - and _ but this is std
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "invalid base64",
			input:   "Not base64!",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			input:   "   \n\t\r  ",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeBase64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("decodeBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name          string
		defaultConfig *ClashConfig
		subscription  *ClashConfig
		want          *ClashConfig
	}{
		{
			name: "subscription overrides non-zero fields",
			defaultConfig: &ClashConfig{
				Port:      7890,
				SocksPort: 7891,
				AllowLan:  true,
				Mode:      "rule",
			},
			subscription: &ClashConfig{
				Port: 9999,
				Mode: "global",
			},
			want: &ClashConfig{
				Port:      9999,
				SocksPort: 7891,
				AllowLan:  true,
				Mode:      "global",
			},
		},
		{
			name: "subscription zero fields do not override",
			defaultConfig: &ClashConfig{
				Port:      7890,
				SocksPort: 7891,
			},
			subscription: &ClashConfig{
				Port: 0,
				Mode: "rule",
			},
			want: &ClashConfig{
				Port:      7890,
				SocksPort: 7891,
				Mode:      "rule",
			},
		},
		{
			name: "empty slice overrides nil slice",
			defaultConfig: &ClashConfig{
				Port:    7890,
				Proxies: nil,
			},
			subscription: &ClashConfig{
				Port:    7890,
				Proxies: []Proxy{},
			},
			want: &ClashConfig{
				Port:    7890,
				Proxies: []Proxy{},
			},
		},
		{
			name: "rule-providers map overrides",
			defaultConfig: &ClashConfig{
				Port: 7890,
				RuleProviders: map[string]RuleProvider{
					"direct": {
						Type:     "http",
						Behavior: "domain",
						URL:      "https://example.com/direct",
						Path:     "./ruleset/direct.list",
						Interval: 604800,
					},
				},
			},
			subscription: &ClashConfig{
				Port: 9999,
				RuleProviders: map[string]RuleProvider{
					"reject": {
						Type:     "http",
						Behavior: "domain",
						URL:      "https://example.com/reject",
						Path:     "./ruleset/reject.list",
						Interval: 604800,
					},
				},
			},
			want: &ClashConfig{
				Port: 9999,
				RuleProviders: map[string]RuleProvider{
					"reject": {
						Type:     "http",
						Behavior: "domain",
						URL:      "https://example.com/reject",
						Path:     "./ruleset/reject.list",
						Interval: 604800,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeConfigs(tt.defaultConfig, tt.subscription)
			if got.Port != tt.want.Port {
				t.Errorf("Port = %v, want %v", got.Port, tt.want.Port)
			}
			if got.SocksPort != tt.want.SocksPort {
				t.Errorf("SocksPort = %v, want %v", got.SocksPort, tt.want.SocksPort)
			}
			if got.RedirPort != tt.want.RedirPort {
				t.Errorf("RedirPort = %v, want %v", got.RedirPort, tt.want.RedirPort)
			}
			if got.MixedPort != tt.want.MixedPort {
				t.Errorf("MixedPort = %v, want %v", got.MixedPort, tt.want.MixedPort)
			}
			if got.AllowLan != tt.want.AllowLan {
				t.Errorf("AllowLan = %v, want %v", got.AllowLan, tt.want.AllowLan)
			}
			if got.BindAddress != tt.want.BindAddress {
				t.Errorf("BindAddress = %v, want %v", got.BindAddress, tt.want.BindAddress)
			}
			if got.Mode != tt.want.Mode {
				t.Errorf("Mode = %v, want %v", got.Mode, tt.want.Mode)
			}
			if got.LogLevel != tt.want.LogLevel {
				t.Errorf("LogLevel = %v, want %v", got.LogLevel, tt.want.LogLevel)
			}
			if got.ExternalController != tt.want.ExternalController {
				t.Errorf("ExternalController = %v, want %v", got.ExternalController, tt.want.ExternalController)
			}
			if len(got.Proxies) != len(tt.want.Proxies) {
				t.Errorf("Proxies length = %v, want %v", len(got.Proxies), len(tt.want.Proxies))
			}
			if len(got.ProxyGroups) != len(tt.want.ProxyGroups) {
				t.Errorf("ProxyGroups length = %v, want %v", len(got.ProxyGroups), len(tt.want.ProxyGroups))
			}
			if len(got.Rules) != len(tt.want.Rules) {
				t.Errorf("Rules length = %v, want %v", len(got.Rules), len(tt.want.Rules))
			}
		})
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	config, err := loadDefaultConfig()
	if err != nil {
		t.Errorf("loadDefaultConfig() error = %v", err)
	}
	if config.Port != 7890 {
		t.Errorf("Port = %v, want 7890", config.Port)
	}
	if config.SocksPort != 7891 {
		t.Errorf("SocksPort = %v, want 7891", config.SocksPort)
	}
	if config.RedirPort != 7892 {
		t.Errorf("RedirPort = %v, want 7892", config.RedirPort)
	}
	if config.AllowLan != true {
		t.Errorf("AllowLan = %v, want true", config.AllowLan)
	}
	if config.Mode != "rule" {
		t.Errorf("Mode = %v, want rule", config.Mode)
	}
	if config.LogLevel != "info" {
		t.Errorf("LogLevel = %v, want info", config.LogLevel)
	}
	if config.RuleProviders == nil {
		t.Errorf("RuleProviders is nil")
	}
	if len(config.RuleProviders) != 4 {
		t.Errorf("RuleProviders length = %v, want 4", len(config.RuleProviders))
	}
	// Check specific providers exist
	if _, ok := config.RuleProviders["direct"]; !ok {
		t.Errorf("RuleProviders missing 'direct'")
	}
	if _, ok := config.RuleProviders["reject"]; !ok {
		t.Errorf("RuleProviders missing 'reject'")
	}
	if _, ok := config.RuleProviders["gfw"]; !ok {
		t.Errorf("RuleProviders missing 'gfw'")
	}
	if _, ok := config.RuleProviders["cncidr"]; !ok {
		t.Errorf("RuleProviders missing 'cncidr'")
	}
	// Check a sample field
	if config.RuleProviders["direct"].Type != "http" {
		t.Errorf("direct.Type = %v, want http", config.RuleProviders["direct"].Type)
	}
	if config.RuleProviders["direct"].Behavior != "domain" {
		t.Errorf("direct.Behavior = %v, want domain", config.RuleProviders["direct"].Behavior)
	}
	if config.RuleProviders["direct"].Format != "mrs" {
		t.Errorf("direct.Format = %v, want mrs", config.RuleProviders["direct"].Format)
	}
}
