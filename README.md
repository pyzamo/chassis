# Chassis

[![Build](https://github.com/pyzamo/chassis-cli/actions/workflows/build.yml/badge.svg)](https://github.com/pyzamo/chassis-cli/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/pyzamo/chassis-cli)](https://github.com/pyzamo/chassis-cli/releases/latest)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](https://go.dev/doc/install)
[![Go Report Card](https://goreportcard.com/badge/github.com/pyzamo/chassis-cli)](https://goreportcard.com/report/github.com/pyzamo/chassis-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight CLI tool to scaffold project directory structures from layout definition files.

## Installation

### From Source

```bash
go install github.com/pyzamo/chassis-cli@latest
```

### Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/pyzamo/chassis-cli/releases) page.

## Quick Start

Create a project structure in seconds:

```bash
# Create a simple layout file
cat > layout.txt << 'EOF'
myapp/
  src/
    main.go
    utils/
      helper.go
  tests/
    main_test.go
  README.md
EOF

# Build the structure
chassis build layout.txt

# Or specify a target directory
chassis build layout.txt ./my-project
```

## Usage

```bash
chassis build <layout-file> <target-dir>
```

### Arguments

- `<layout-file>` - Path to layout definition file (or `-` for stdin)
- `<target-dir>` - Target directory (defaults to current directory)

### Options

- `-v, --verbose` - Print every path created/skipped
- `--indent <int>` - Expected space width for plain-text parser (default: 2)

## Layout Formats

### Plain-Text Tree

```
project/
  src/
    main.py
    utils/
      helpers.py
      constants.py
  tests/
    test_main.py
  requirements.txt
  README.md
```

### YAML

Use `null` for files, `{}` for empty directories

```yaml
backend:
  cmd:
    server:
      main.go: null
  internal:
    handlers: {}
    models: {}
  go.mod: null
frontend:
  src:
    App.tsx: null
    index.tsx: null
  package.json: null
```

### JSON
```json
{
  "webapp": {
    "client": {
      "src": {
        "App.tsx": null,
        "index.tsx": null
      },
      "package.json": null
    },
    "server": {
      "main.js": null,
      "package.json": null
    }
  }
}
```

## Examples

### Basic Usage

```bash
# From a file
chassis build layout.txt my-project

# From stdin
cat layout.yaml | chassis build - my-project

# With verbose output
chassis build -v structure.json app
```

### Real-World Examples

**Go Microservice**
```yaml
api:
  cmd:
    server:
      main.go: null
  internal:
    handlers:
      auth.go: null
      user.go: null
    models:
      user.go: null
  go.mod: null
  Dockerfile: null
  README.md: null
```

Save as `microservice.yaml` and run:
```bash
chassis build microservice.yaml my-service
```

**Python Package**
```json
{
  "my_package": {
    "src": {
      "my_package": {
        "__init__.py": null,
        "core.py": null,
        "utils.py": null
      }
    },
    "tests": {
      "test_core.py": null,
      "test_utils.py": null
    },
    "setup.py": null,
    "requirements.txt": null,
    "README.md": null
  }
}
```

Save as `python-package.json` and run:
```bash
chassis build python-package.json
```
