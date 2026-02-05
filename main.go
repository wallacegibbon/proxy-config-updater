package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type ClashConfig struct {
	Port               int                     `yaml:"port"`
	SocksPort          int                     `yaml:"socks-port"`
	RedirPort          int                     `yaml:"redir-port"`
	MixedPort          int                     `yaml:"mixed-port"`
	AllowLan           bool                    `yaml:"allow-lan"`
	BindAddress        string                  `yaml:"bind-address"`
	Mode               string                  `yaml:"mode"`
	LogLevel           string                  `yaml:"log-level"`
	ExternalController string                  `yaml:"external-controller"`
	Proxies            []Proxy                 `yaml:"proxies"`
	ProxyGroups        []ProxyGroup            `yaml:"proxy-groups"`
	Rules              []string                `yaml:"rules"`
	RuleProviders      map[string]RuleProvider `yaml:"rule-providers,omitempty"`
}

type Proxy struct {
	Name           string         `yaml:"name"`
	Type           string         `yaml:"type"`
	Server         string         `yaml:"server"`
	Port           int            `yaml:"port"`
	Password       string         `yaml:"password,omitempty"`
	UUID           string         `yaml:"uuid,omitempty"`
	Cipher         string         `yaml:"cipher,omitempty"`
	Network        string         `yaml:"network,omitempty"`
	UDP            bool           `yaml:"udp,omitempty"`
	TLS            bool           `yaml:"tls,omitempty"`
	SkipCertVerify bool           `yaml:"skip-cert-verify,omitempty"`
	Extra          map[string]any `yaml:",inline,omitempty"`
}

type ProxyGroup struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies"`
	URL      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

type RuleProvider struct {
	Type     string `yaml:"type"`
	Behavior string `yaml:"behavior"`
	Format   string `yaml:"format,omitempty"`
	URL      string `yaml:"url"`
	Path     string `yaml:"path"`
	Interval int    `yaml:"interval"`
}

func main() {
	var output string
	var pretty bool = true
	urlFile := ""

	// Manual flag parsing to allow flags anywhere
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 0 && arg[0] == '-' {
			switch arg {
			case "-output":
				if i+1 < len(args) {
					output = args[i+1]
					i++
				} else {
					fmt.Fprintln(os.Stderr, "Error: -output requires a value")
					os.Exit(1)
				}
			case "-pretty":
				pretty = true
			case "-pretty=false":
				pretty = false
			default:
				fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
				os.Exit(1)
			}
		} else if urlFile == "" {
			urlFile = arg
		} else {
			fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", arg)
			os.Exit(1)
		}
	}

	if urlFile == "" {
		fmt.Fprintln(os.Stderr, "Usage: base64-subscription-config <url-file> [options]")
		fmt.Fprintln(os.Stderr, "  -output string   Output file path (default: stdout)")
		fmt.Fprintln(os.Stderr, "  -pretty          Pretty print output (default true)")
		os.Exit(1)
	}
	urlBytes, err := os.ReadFile(urlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading URL file: %v\n", err)
		os.Exit(1)
	}

	subscriptionURL := strings.TrimSpace(string(urlBytes))
	if subscriptionURL == "" {
		fmt.Fprintln(os.Stderr, "Error: URL file is empty")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Reading URL from file: %s\n", urlFile)
	fmt.Fprintf(os.Stderr, "Fetching subscription from: %s\n", subscriptionURL)
	encodedContent, err := fetchContent(subscriptionURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching content: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Decoding base64 content...")
	decodedContent, err := decodeBase64(encodedContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Content is not base64 encoded, trying raw YAML...\n")
		decodedContent = encodedContent
	}

	fmt.Fprintln(os.Stderr, "Parsing Clash configuration...")
	config, err := parseClashConfig(decodedContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Configuration loaded: %d proxies, %d groups, %d rules\n",
		len(config.Proxies), len(config.ProxyGroups), len(config.Rules))

	var outputWriter io.Writer = os.Stdout
	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		outputWriter = file
		fmt.Fprintf(os.Stderr, "Writing configuration to: %s\n", output)
	}

	if pretty {
		// Load default configuration
		defaultConfig, err := loadDefaultConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: unable to load default config: %v\n", err)
			defaultConfig = &ClashConfig{}
		}

		// Merge default config with subscription config
		mergedConfig := mergeConfigs(defaultConfig, config)

		yamlData, err := yaml.Marshal(mergedConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
			os.Exit(1)
		}
		outputWriter.Write(yamlData)
	} else {
		outputWriter.Write([]byte(decodedContent))
	}

	fmt.Fprintln(os.Stderr, "\nDone!")
}

func fetchContent(url string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "ClashSubscriptionParser/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func decodeBase64(encoded string) (string, error) {
	// Remove UTF-8 BOM if present
	encodedBytes := []byte(encoded)
	if bytes.HasPrefix(encodedBytes, []byte{0xEF, 0xBB, 0xBF}) {
		encodedBytes = encodedBytes[3:]
	}

	// Remove common data URL prefixes
	encodedStr := string(encodedBytes)
	prefixes := []string{
		"data:application/octet-stream;base64,",
		"data:text/plain;base64,",
		"data:application/x-yaml;base64,",
		"data:;base64,",
		"base64,",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(encodedStr, prefix) {
			encodedStr = encodedStr[len(prefix):]
			break
		}
	}

	// Remove all whitespace characters (spaces, tabs, newlines, carriage returns)
	cleaned := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, strings.TrimSpace(encodedStr))

	// Try different base64 encodings
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	var firstErr error
	for _, enc := range encodings {
		decoded, err := enc.DecodeString(cleaned)
		if err == nil {
			return string(decoded), nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	return "", firstErr
}

func parseClashConfig(content string) (*ClashConfig, error) {
	var config ClashConfig

	err := yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func loadDefaultConfig() (*ClashConfig, error) {
	return &ClashConfig{
		Port:      7890,
		SocksPort: 7891,
		RedirPort: 7892,
		AllowLan:  true,
		Mode:      "rule",
		LogLevel:  "info",
		RuleProviders: map[string]RuleProvider{
			"direct": {
				Type:     "http",
				Behavior: "domain",
				Format:   "mrs",
				URL:      "https://edgeone.gh-proxy.org/https://github.com/DustinWin/ruleset_geodata/releases/download/mihomo-ruleset/cn-lite.mrs",
				Path:     "./ruleset/direct.list",
				Interval: 604800,
			},
			"reject": {
				Type:     "http",
				Behavior: "domain",
				Format:   "mrs",
				URL:      "https://edgeone.gh-proxy.org/raw.githubusercontent.com/privacy-protection-tools/anti-ad.github.io/master/docs/mihomo.mrs",
				Path:     "./ruleset/aiti-ad.list",
				Interval: 604800,
			},
			"gfw": {
				Type:     "http",
				Behavior: "domain",
				URL:      "https://edgeone.gh-proxy.org/raw.githubusercontent.com/Loyalsoldier/clash-rules/release/gfw.txt",
				Path:     "./ruleset/gfw.list",
				Interval: 604800,
			},
			"cncidr": {
				Type:     "http",
				Behavior: "ipcidr",
				Format:   "mrs",
				URL:      "https://edgeone.gh-proxy.org/https://github.com/DustinWin/ruleset_geodata/releases/download/mihomo-ruleset/cnip.mrs",
				Path:     "./ruleset/cncidr.list",
				Interval: 604800,
			},
		},
	}, nil
}

func mergeConfigs(defaultConfig, subscriptionConfig *ClashConfig) *ClashConfig {
	// Start with default config
	merged := *defaultConfig

	// Overwrite with subscription config fields if they are non-zero
	// For simplicity, we'll just unmarshal subscription into default
	// But we need to preserve default values for fields not in subscription
	// We'll use reflection to copy non-zero fields from subscription to default

	subVal := reflect.ValueOf(subscriptionConfig).Elem()
	defVal := reflect.ValueOf(&merged).Elem()

	for i := 0; i < subVal.NumField(); i++ {
		subField := subVal.Field(i)
		defField := defVal.Field(i)

		// Check if the field is zero in subscription config
		// If not zero, copy to default
		if !subField.IsZero() {
			defField.Set(subField)
		}
	}

	return &merged
}
