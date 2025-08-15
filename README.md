# Chassis

[![Build](https://github.com/pyzamo/chassis-cli/actions/workflows/build.yml/badge.svg)](https://github.com/pyzamo/chassis-cli/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/pyzamo/chassis-cli)](https://github.com/pyzamo/chassis-cli/releases/latest)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](https://go.dev/doc/install)
[![Go Report Card](https://goreportcard.com/badge/github.com/pyzamo/chassis-cli)](https://goreportcard.com/report/github.com/pyzamo/chassis-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Scaffold project directory structures from layout definition files.

## Installation

```bash
go install github.com/pyzamo/chassis@latest
```

Or download from [releases](https://github.com/pyzamo/chassis-cli/releases).

## Usage

```bash
# Build from layout file
chassis build layout.txt my-project

# Analyze existing project (requires GEMINI_API_KEY env var)
chassis analyze ./existing-project > template.txt
chassis build template.txt new-project

# From stdin
echo -e "src/\n  main.go" | chassis build - .
```

## Commands

### build
Creates directory structure from a layout file.

```bash
chassis build <layout-file> [target-dir]
```

- Supports plain text (indented), YAML, and JSON formats
- Auto-detects format from file extension
- Skips existing files/directories

### analyze
Extracts project structure into a reusable template using AI.

```bash
chassis analyze <source> [--format tree|yaml|json] [--max-depth N]
```

- Works with local directories or GitHub repos
- Filters out common artifacts (node_modules, .git, etc.)
- Requires `GEMINI_API_KEY` environment variable

## Layout Formats

### Plain Text
```
project/
  src/
    main.go
    utils/
  tests/
  go.mod
```

### YAML
```yaml
project:
  src:
    main.go: null
    utils: {}
  tests: {}
  go.mod: null
```

### JSON
```json
{
  "project": {
    "src": {
      "main.go": null,
      "utils": {}
    },
    "tests": {},
    "go.mod": null
  }
}
```

## Example Workflow

```bash
# Analyze and recreate a Go project structure
export GEMINI_API_KEY='your-key'  # Get from https://aistudio.google.com/app/apikey
chassis analyze https://github.com/gin-gonic/gin --max-depth 2 > gin.txt
chassis build gin.txt my-gin-app

# Quick scaffolding
cat > layout.yaml << EOF
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
  package.json: null
EOF
chassis build layout.yaml my-fullstack-app
```

## License

MIT