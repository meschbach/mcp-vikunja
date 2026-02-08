# MCP Vikunja

A Model Context Protocol (MCP) server for Vikunja task management integration.

## Overview

This project provides an MCP server that allows LLMs and AI assistants to interact with Vikunja task management through standardized tool calls.

## Features

- List tasks from Vikunja
- Get task details
- Create new tasks
- Update existing tasks

## Installation

```bash
go get github.com/meschbach/mcp-vikunja
```

## Usage

### Building

```bash
make build
```

### Running

```bash
make run
# or
./bin/mcp-vikunja
```

### Environment Variables

- `VIKUNJA_API_TOKEN` - Your Vikunja API token (required)
- `VIKUNJA_BASE_URL` - Vikunja instance URL (optional, defaults to https://vikunja.io/api/v1)

## Development

See [AGENTS.md](AGENTS.md) for coding guidelines and commands.

## License

MIT