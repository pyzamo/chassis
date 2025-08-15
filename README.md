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
go install github.com/pyzamo/chassis@latest
```

### Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/pyzamo/chassis-cli/releases) page.

## Commands

### `chassis analyze` - Extract Templates with AI

Analyzes existing projects and uses Google Gemini AI to generate reusable scaffolding templates.

```bash
chassis analyze <source> [flags]
```

**Arguments:**
- `<source>` - Local directory or GitHub repo URL

**Flags:**
- `--format string` - Output format: tree, yaml, json (default: tree)
- `--max-depth int` - Maximum depth to analyze (default: 5)

**Examples:**
```bash
# Analyze local project
chassis analyze ./my-app > template.txt

# Analyze GitHub repository
chassis analyze https://github.com/gin-gonic/gin > gin-template.txt
```

### `chassis build` - Create Project Structure

Creates directory structure from a layout definition file.

```bash
chassis build <layout-file> [target-dir]
```

**Arguments:**
- `<layout-file>` - Path to layout file (or `-` for stdin)
- `[target-dir]` - Target directory (defaults to current directory)

**Flags:**
- `-v, --verbose` - Print every path created/skipped
- `--indent int` - Space width for plain-text parser (default: 2)

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

## Quick Start

```bash
# 1. Set up Gemini API key (one-time, get free key at https://aistudio.google.com/app/apikey)
export GEMINI_API_KEY='your-key-here'

# 2. Analyze existing project to extract template
chassis analyze ./my-project > template.txt

# 3. Create new projects from template
chassis build template.txt ./new-project
```

## Examples

### Complete Workflow

```bash
# Extract template from an existing Express.js app
chassis analyze ./my-express-app > express-template.txt

# app/
#   # Core application files
#   server.js
#   app.js
#   # Route definitions
#   routes/
#     api.js
#   # Data models
#   models/
#     model.js
#   # Middleware functions
#   middleware/
#     auth.js
#   # Configuration
#   config/
#     config.js
#   package.json

# Create new projects from this template
chassis build express-template.txt project-1
```

### Quick Examples

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

Contributions are welcome! Please feel free to submit a Pull Request.