# Contributing to MCP Brasil

Thank you for your interest in contributing to MCP Brasil!

## How to Contribute

### Reporting Bugs

- Check if the bug has already been reported in [Issues](https://github.com/anderson-ufrj/mcp-brasil/issues)
- If not, create a new issue with a clear title and description
- Include steps to reproduce, expected behavior, and actual behavior

### Suggesting Features

- Open an issue describing the feature
- Explain why this feature would be useful
- If possible, suggest an implementation approach

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests to ensure everything works
5. Commit your changes following [Conventional Commits](https://www.conventionalcommits.org/)
6. Push to your fork (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Commit Message Format

```
type(scope): description

[optional body]
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

Examples:
- `feat(tools): add search by CNPJ`
- `fix(client): handle API timeout`
- `docs: update README with examples`

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/mcp-brasil.git
cd mcp-brasil

# Install dependencies
go mod tidy

# Build
go build -o mcp-brasil ./cmd/server

# Run tests
go test ./...
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for exported functions

## Adding New Tools

To add a new Portal da Transparencia endpoint:

1. Add the API method in `pkg/transparencia/client.go`
2. Add the tool definition in `cmd/server/main.go`
3. Update the README with the new tool
4. Add tests if applicable

## Questions?

Feel free to open an issue for any questions about contributing.
