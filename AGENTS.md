# AGENTS.md

This file contains essential information for working effectively in this codebase.

## Project Overview

This is a Go CLI application that parses Clash subscription configurations. It reads a subscription URL from a file, fetches base64 encoded content from that URL, decodes it, parses YAML, and outputs the configuration with a summary.

## Essential Commands

### Build
```bash
go build
```

This compiles the application to an executable named `base64-subscription-config` (the module name).

### Run
```bash
./base64-subscription-config <url-file> [options]
```

The `<url-file>` argument is a file containing the subscription URL (one URL per file).

Options:
- `-output <path>` - Output file path (default: stdout)
- `-pretty` - Pretty print output (default: true)

Examples:
```bash
# Create a file with the subscription URL
echo "https://example.com/subscription" > url.txt

# Parse and output to stdout
./base64-subscription-config url.txt

# Save to file
./base64-subscription-config url.txt -output config.yaml

# Raw output (not pretty printed)
./base64-subscription-config url.txt -pretty=false
```

### Go Module Operations
```bash
go mod tidy     # Clean up dependencies
go mod download # Download dependencies
```

## Code Organization

This is a single-file application (`main.go`) with a straightforward structure:

- **Entry Point**: `main()` function handles CLI flags, reads URL from file, orchestrates the workflow
- **Data Models**: Three structs define the Clash configuration schema:
  - `ClashConfig` - Top-level configuration container
  - `Proxy` - Individual proxy server configuration
  - `ProxyGroup` - Group of proxies with selection strategy
- **Core Functions**:
  - `os.ReadFile(filename)` - Read subscription URL from file
  - `fetchContent(url)` - HTTP GET with 30s timeout, custom User-Agent
  - `decodeBase64(encoded)` - Base64 decoding with whitespace trimming
  - `parseClashConfig(content)` - YAML unmarshaling into config struct
  - `loadDefaultConfig()` - Returns embedded default Clash configuration values
  - `mergeConfigs(default, subscription)` - Merges subscription config into default using reflection

## Code Patterns and Conventions

### Error Handling
- Errors are printed to stderr using `fmt.Fprintln(os.Stderr, ...)`
- All errors result in `os.Exit(1)` after printing
- No error recovery - failures are terminal

### Output Convention
- **Progress/status messages**: stderr (file reading status, fetching status, decoding status, summary)
- **Configuration output**: stdout or file (YAML content)
- **Default behavior**: Pretty print YAML, output to stdout
- **Status messages include**: URL filename being read, actual subscription URL being fetched

### Naming Conventions
- Struct names use PascalCase: `ClashConfig`, `Proxy`, `ProxyGroup`
- Field names use PascalCase for exported fields
- YAML tags match Clash configuration schema (kebab-case: `socks-port`, `proxy-groups`)
- Function names use camelCase: `fetchContent`, `decodeBase64`, `parseClashConfig`

### YAML Struct Tags
All struct fields use `yaml:"field-name"` tags for unmarshaling. Note:
- Fields with `omitempty` are optional and omitted when zero/empty
- Proxy struct has `,inline` tag for extra fields: `yaml:",inline,omitempty"`

### HTTP Client Pattern
The HTTP client is created with:
- 30 second timeout
- Custom User-Agent header: `"ClashSubscriptionParser/1.0"`
- Standard error handling for non-200 status codes

## Dependencies

- **Go version**: 1.25.6
- **gopkg.in/yaml.v3** - YAML parsing and generation

No other external dependencies.

## Testing

Test files exist (`main_test.go`) with comprehensive unit tests for:
- Base64 decoding with various edge cases
- Configuration merging logic (`mergeConfigs`)
- Default configuration loading (`loadDefaultConfig`)

Run tests with: `go test -v`

## Gotchas and Non-Obvious Patterns

1. **File-based URL Input**: The tool accepts a filename, not a URL directly. The file must contain the subscription URL (trailing whitespace is automatically trimmed). Empty files cause an error.

2. **Flag Parameter Name**: Use `-output` not `-out` or `-o` for the output file parameter

3. **Pretty Print Default**: The `-pretty` flag defaults to `true`. Set to `false` for raw YAML output.

4. **Output Separation**: Progress and error messages go to stderr; configuration data goes to stdout/file. This allows piping configuration while still seeing status messages.

5. **User-Agent**: All HTTP requests use a custom User-Agent header for identification.

6. **Base64 Trimming**: The `decodeBase64` function trims whitespace before decoding to handle common formatting issues.

7. **Inline Extra Fields**: The `Proxy` struct uses `yaml:",inline,omitempty"` to capture any additional fields not explicitly defined, providing flexibility for proxy variations.

8. **Binary Already Present**: The directory may contain a compiled binary (`base64-subscription-config`). Rebuild after making code changes.

9. **YAML Unmarshaling**: The code silently ignores unknown YAML fields during unmarshaling (standard behavior of yaml.v3).

10. **Default Configuration Merging**: When pretty printing (`-pretty=true`), the tool merges subscription configuration with embedded default values (port: 7890, socks-port: 7891, redir-port: 7892, allow-lan: true, mode: rule, log-level: info). Subscription fields override defaults. Default rule providers are also included (direct, reject, gfw, cncidr) unless the subscription provides its own rule-providers map.

11. **Flexible Flag Order**: Flags can appear anywhere in the command line (before or after the URL file argument). Example: `./base64-subscription-config url.txt -output config.yaml` or `./base64-subscription-config -output config.yaml url.txt`

12. **Rule Providers Map Replacement**: If a subscription includes a `rule-providers` map, it completely replaces the default rule providers map (the entire map is replaced, not merged). This is because maps are treated as single values during configuration merging.

## Project Structure

```
.
├── main.go                      # Single source file containing all logic
├── main_test.go                 # Unit tests for core functionality
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── README.md                    # User documentation
└── base64-subscription-config    # Compiled binary (after build)
```

No subdirectories, Makefiles, build scripts, or CI/CD configuration.
