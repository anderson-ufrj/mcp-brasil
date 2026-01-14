# MCP Brasil

MCP (Model Context Protocol) server for Brazilian government transparency data.

## Features

- **search_contracts**: Search federal government contracts
- **search_servidores**: Search federal public servants by name
- **get_remuneracao**: Get salary data for a public servant by CPF
- **search_convenios**: Search government agreements by state
- **search_ceis**: Search sanctioned companies (CEIS)
- **list_orgaos**: List known government organization codes

## Data Source

All data comes from [Portal da Transparência](https://api.portaldatransparencia.gov.br).

## Installation

```bash
# Clone the repository
git clone https://github.com/anderson-ufrj/mcp-brasil.git
cd mcp-brasil

# Build
go build -o mcp-brasil ./cmd/server
```

## Configuration

Get an API key from [Portal da Transparência](https://api.portaldatransparencia.gov.br/) and set it:

```bash
export TRANSPARENCY_API_KEY=your_api_key_here
```

## Usage with Claude Code

Add to your Claude Code settings (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "mcp-brasil": {
      "command": "/full/path/to/mcp-brasil",
      "env": {
        "TRANSPARENCY_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

## Usage with Claude Desktop

Add to your Claude Desktop config:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "mcp-brasil": {
      "command": "/full/path/to/mcp-brasil",
      "env": {
        "TRANSPARENCY_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

## Example Queries

Once configured, you can ask Claude:

- "Search for government contracts from the Ministry of Health"
- "Find public servants named 'Maria Silva'"
- "What are the federal agreements in Minas Gerais?"
- "Check if company CNPJ 12.345.678/0001-90 is sanctioned"

## License

MIT
