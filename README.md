# MCP Brasil

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-Compatible-green.svg)](https://modelcontextprotocol.io/)

MCP (Model Context Protocol) server for Brazilian government transparency data.

## Features

| Tool | Description |
|------|-------------|
| `search_contracts` | Search federal government contracts |
| `search_servidores` | Search federal public servants by name |
| `get_remuneracao` | Get salary data for a public servant by CPF |
| `search_convenios` | Search government agreements by state |
| `search_ceis` | Search sanctioned companies (CEIS) |
| `list_orgaos` | List known government organization codes |

## Data Source

All data comes from [Portal da Transparencia](https://api.portaldatransparencia.gov.br) - the official Brazilian government transparency portal.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/anderson-ufrj/mcp-brasil.git
cd mcp-brasil

# Build
go build -o mcp-brasil ./cmd/server
```

### Requirements

- Go 1.22 or later
- API key from Portal da Transparencia (optional, but recommended)

## Configuration

1. Get an API key from [Portal da Transparencia](https://api.portaldatransparencia.gov.br/)
2. Set the environment variable:

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

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |

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
- "List available government organization codes"

## Known Organization Codes

| Code | Organization |
|------|--------------|
| 36000 | Ministerio da Saude |
| 26000 | Ministerio da Educacao |
| 25000 | Ministerio da Economia |
| 30000 | Ministerio da Justica |
| 52000 | Ministerio da Defesa |
| 35000 | Ministerio das Relacoes Exteriores |
| 44000 | Ministerio do Meio Ambiente |

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Citation

If you use this software in your research, please cite it:

```bibtex
@software{mcp_brasil,
  author = {Silva, Anderson Henrique},
  title = {MCP Brasil},
  year = {2025},
  url = {https://github.com/anderson-ufrj/mcp-brasil}
}
```

## Related Projects

- [langchain-maritaca](https://github.com/anderson-ufrj/langchain-maritaca) - LangChain integration for Maritaca AI
- [Model Context Protocol](https://modelcontextprotocol.io/) - Official MCP specification
