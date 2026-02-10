# Proxy Config Updater

A command line tool in Go that parses Clash subscription URLs (which contain base64 encoded content) and generates Clash configurations.

## Features

- Read subscription URL from a file
- Fetch and decode base64 encoded Clash subscription content
- Parse and validate Clash YAML configurations
- Pretty print or raw output
- Output to file or stdout
- Displays summary of parsed configuration (proxies, groups, rules count)
- Merges subscription configuration with default values (ports, mode, log-level, rule providers)

## Installation

```bash
go build
```

## Usage

```bash
./proxy-config-updater <url-file> [options]
```

### Options

- `-output <path>` - Output file path (default: stdout)
- `-pretty` - Pretty print output (default: true)

### Examples

Create a file containing the subscription URL:
```bash
echo "https://example.com/subscription" > url.txt
```

Parse subscription and output to stdout:
```bash
./proxy-config-updater url.txt
```

Save to file:
```bash
./proxy-config-updater url.txt -output config.yaml
```

Output raw (not pretty printed):
```bash
./proxy-config-updater url.txt -pretty=false
```

## Default Configuration

When using pretty print mode (default), the tool merges subscription configuration with default values:

- Port: 7890
- Socks Port: 7891
- Redir Port: 7892
- Allow LAN: true
- Mode: rule
- Log Level: info
- Rule Providers: Four default rule providers are included:
  - `direct`: Domain-based rules for direct connections
  - `reject`: Domain-based rules for ad blocking
  - `gfw`: Domain-based rules for GFW list
  - `cncidr`: IP CIDR rules for Chinese IP ranges

Fields present in the subscription configuration override these defaults. Rule providers from the subscription replace the default rule providers entirely (the default rule providers map is replaced if the subscription includes any `rule-providers` section).

## Dependencies

- Go 1.16+
- gopkg.in/yaml.v3

## License

MIT
