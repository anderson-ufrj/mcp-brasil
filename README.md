# MCP Brasil

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-Compatible-green.svg)](https://modelcontextprotocol.io/)
[![Version](https://img.shields.io/badge/Version-2.0.0-brightgreen.svg)](https://github.com/anderson-ufrj/mcp-brasil/releases)

MCP (Model Context Protocol) server for Brazilian government and public data. Provides unified access to 17 tools across 5 official Brazilian APIs.

## Data Sources

| Source | Description | Tools |
|--------|-------------|-------|
| **Portal da Transparencia** | Federal government transparency data | 6 |
| **IBGE** | Brazilian geography and demographics | 3 |
| **Minha Receita** | Company (CNPJ) lookup | 1 |
| **Banco Central** | Economic indicators and exchange rates | 4 |
| **PNCP** | Public procurement contracts | 3 |

## Tools (17 total)

### Portal da Transparencia

| Tool | Description |
|------|-------------|
| `search_contracts` | Search federal government contracts |
| `search_servidores` | Search federal public servants by name |
| `get_remuneracao` | Get salary data for a public servant by CPF |
| `search_convenios` | Search government agreements by state |
| `search_ceis` | Search sanctioned companies (CEIS) |
| `list_orgaos` | List known government organization codes |

### IBGE (Geography & Demographics)

| Tool | Description |
|------|-------------|
| `ibge_states` | List all Brazilian states with region info |
| `ibge_municipalities` | List municipalities (optionally by state) |
| `ibge_population` | Get population data for a location |

### Minha Receita (CNPJ)

| Tool | Description |
|------|-------------|
| `cnpj_lookup` | Get company data by CNPJ (address, activities, partners) |

### Banco Central (BCB)

| Tool | Description |
|------|-------------|
| `bcb_selic` | Get SELIC interest rate history |
| `bcb_ipca` | Get IPCA inflation rate history |
| `bcb_exchange_rate` | Get currency exchange rates (USD, EUR, etc.) |
| `bcb_indicator` | Get any BCB economic indicator by code |

### PNCP (Public Procurement)

| Tool | Description |
|------|-------------|
| `pncp_contracts` | Search public procurement publications |
| `pncp_price_registrations` | Search price registration records |
| `pncp_modalities` | List procurement modality codes |

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
- API key from Portal da Transparencia (optional, for transparency tools)

## Configuration

1. Get an API key from [Portal da Transparencia](https://api.portaldatransparencia.gov.br/) (only needed for transparency tools)
2. Set the environment variable:

```bash
export TRANSPARENCY_API_KEY=your_api_key_here
```

**Note**: IBGE, CNPJ, BCB, and PNCP tools work without authentication.

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

**Government Transparency**
- "Search for government contracts from the Ministry of Health"
- "Find public servants named 'Maria Silva'"
- "What are the federal agreements in Minas Gerais?"
- "Check if company CNPJ 12.345.678/0001-90 is sanctioned"

**Geography & Demographics**
- "List all Brazilian states"
- "What municipalities are in Sao Paulo?"
- "What is the population of Belo Horizonte?"

**Company Data**
- "Look up CNPJ 00.000.000/0001-91"
- "Who are the partners of Petrobras?"

**Economic Indicators**
- "What is the current SELIC rate?"
- "Show me IPCA inflation for the last 12 months"
- "What is the USD to BRL exchange rate today?"

**Public Procurement**
- "Search for recent electronic auctions in SP"
- "Find price registrations for IT services"

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

## Economic Indicators (BCB)

| Indicator | Series Code | Description |
|-----------|-------------|-------------|
| `selic` | 11 | SELIC daily rate |
| `selic_monthly` | 4390 | SELIC accumulated monthly |
| `ipca` | 433 | IPCA monthly inflation |
| `igpm` | 189 | IGP-M monthly |
| `cdi` | 12 | CDI daily rate |

## Procurement Modalities (PNCP)

| Modality | Code |
|----------|------|
| Pregao Eletronico | 6 |
| Concorrencia Eletronica | 1 |
| Concorrencia | 2 |
| Concurso | 3 |
| Leilao Eletronico | 4 |
| Leilao | 5 |
| Dialogo Competitivo | 7 |
| Credenciamento | 8 |

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
  version = {2.0.0},
  url = {https://github.com/anderson-ufrj/mcp-brasil}
}
```

## Related Projects

- [langchain-maritaca](https://github.com/anderson-ufrj/langchain-maritaca) - LangChain integration for Maritaca AI
- [Cidadao.AI](https://github.com/anderson-ufrj/cidadao.ai-backend) - Multi-agent transparency platform
- [Model Context Protocol](https://modelcontextprotocol.io/) - Official MCP specification
