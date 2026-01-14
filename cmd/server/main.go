// MCP Brasil - Model Context Protocol server for Brazilian government data.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anderson-ufrj/mcp-brasil/pkg/transparencia"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var client *transparencia.Client

func main() {
	// Get API key from environment
	apiKey := os.Getenv("TRANSPARENCY_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Warning: TRANSPARENCY_API_KEY not set, some features may not work")
	}

	// Create Portal da Transparencia client
	client = transparencia.NewClient(apiKey)

	// Create MCP server
	s := server.NewMCPServer(
		"MCP Brasil",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(true, false),
	)

	// Register tools
	registerTools(s)

	// Register resources
	registerResources(s)

	// Run server over stdio
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func registerTools(s *server.MCPServer) {
	// Tool: search_contracts
	contractsTool := mcp.NewTool("search_contracts",
		mcp.WithDescription("Search government contracts from Portal da Transparencia. Returns federal government contracts filtered by organization code."),
		mcp.WithString("orgao_code",
			mcp.Description("Organization SIAPE code (e.g. 36000 for Ministry of Health). Use list_orgaos to see available codes."),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default 1)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Results per page (max 500, default 100)"),
		),
	)
	s.AddTool(contractsTool, handleSearchContracts)

	// Tool: search_servidores
	servidoresTool := mcp.NewTool("search_servidores",
		mcp.WithDescription("Search federal public servants by name. Returns information about civil servants including their organization and position."),
		mcp.WithString("nome",
			mcp.Required(),
			mcp.Description("Name of the public servant to search"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default 1)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Results per page (max 500, default 100)"),
		),
	)
	s.AddTool(servidoresTool, handleSearchServidores)

	// Tool: get_remuneracao
	remuneracaoTool := mcp.NewTool("get_remuneracao",
		mcp.WithDescription("Get salary data for a public servant by CPF. Returns detailed remuneration including base salary, bonuses, and deductions."),
		mcp.WithString("cpf",
			mcp.Required(),
			mcp.Description("CPF of the public servant (11 digits, numbers only)"),
		),
		mcp.WithString("mes_ano",
			mcp.Description("Month/Year in MM/YYYY format (e.g. 01/2024). Defaults to last month."),
		),
	)
	s.AddTool(remuneracaoTool, handleGetRemuneracao)

	// Tool: search_convenios
	conveniosTool := mcp.NewTool("search_convenios",
		mcp.WithDescription("Search federal government agreements (convenios) by state. Returns information about agreements, including values and status."),
		mcp.WithString("uf",
			mcp.Description("State code (e.g. MG, SP, RJ). Defaults to MG."),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default 1)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Results per page (max 500, default 100)"),
		),
	)
	s.AddTool(conveniosTool, handleSearchConvenios)

	// Tool: search_ceis
	ceisTool := mcp.NewTool("search_ceis",
		mcp.WithDescription("Search sanctioned companies in CEIS (Cadastro de Empresas Inidoneas e Suspensas). Returns companies that are banned from government contracts."),
		mcp.WithString("cnpj",
			mcp.Description("CNPJ of the company to search (optional, 14 digits)"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number (default 1)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Results per page (max 500, default 100)"),
		),
	)
	s.AddTool(ceisTool, handleSearchCEIS)

	// Tool: list_orgaos
	orgaosTool := mcp.NewTool("list_orgaos",
		mcp.WithDescription("List known government organization codes (SIAPE). Use these codes with other tools like search_contracts."),
	)
	s.AddTool(orgaosTool, handleListOrgaos)
}

func registerResources(s *server.MCPServer) {
	// Resource: API documentation
	docResource := mcp.NewResource(
		"docs://api-reference",
		"API Reference",
		mcp.WithResourceDescription("Documentation for Brazilian government transparency APIs"),
		mcp.WithMIMEType("text/markdown"),
	)
	s.AddResource(docResource, handleDocResource)
}

// Tool handlers

func handleSearchContracts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	orgaoCode, _ := request.GetArguments()["orgao_code"].(string)
	page := getIntArg(request, "page", 1)
	pageSize := getIntArg(request, "page_size", 100)

	result, err := client.SearchContracts(ctx, orgaoCode, page, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching contracts: %v", err)), nil
	}

	return toJSONResult(result)
}

func handleSearchServidores(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nome, err := request.RequireString("nome")
	if err != nil {
		return mcp.NewToolResultError("Parameter 'nome' is required"), nil
	}

	page := getIntArg(request, "page", 1)
	pageSize := getIntArg(request, "page_size", 100)

	result, err := client.SearchServidores(ctx, nome, page, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching servidores: %v", err)), nil
	}

	return toJSONResult(result)
}

func handleGetRemuneracao(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cpf, err := request.RequireString("cpf")
	if err != nil {
		return mcp.NewToolResultError("Parameter 'cpf' is required"), nil
	}

	mesAno, _ := request.GetArguments()["mes_ano"].(string)

	result, err := client.GetServidorRemuneracao(ctx, cpf, mesAno)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting remuneracao: %v", err)), nil
	}

	return toJSONResult(result)
}

func handleSearchConvenios(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uf, _ := request.GetArguments()["uf"].(string)
	page := getIntArg(request, "page", 1)
	pageSize := getIntArg(request, "page_size", 100)

	result, err := client.SearchConvenios(ctx, uf, page, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching convenios: %v", err)), nil
	}

	return toJSONResult(result)
}

func handleSearchCEIS(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cnpj, _ := request.GetArguments()["cnpj"].(string)
	page := getIntArg(request, "page", 1)
	pageSize := getIntArg(request, "page_size", 100)

	result, err := client.SearchCEIS(ctx, cnpj, page, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching CEIS: %v", err)), nil
	}

	return toJSONResult(result)
}

func handleListOrgaos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result := client.ListOrgaos()
	return toJSONResult(result)
}

// Resource handlers

func handleDocResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "docs://api-reference",
			MIMEType: "text/markdown",
			Text:     getAPIDocumentation(),
		},
	}, nil
}

// Helper functions

func getIntArg(request mcp.CallToolRequest, key string, defaultVal int) int {
	args := request.GetArguments()
	if val, ok := args[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

func toJSONResult(data interface{}) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error encoding result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func getAPIDocumentation() string {
	orgaos := transparencia.KnownOrgaos
	orgaosJSON, _ := json.MarshalIndent(orgaos, "", "  ")

	return fmt.Sprintf(`# MCP Brasil - API Reference

## Overview
This MCP server provides access to Brazilian government transparency data from Portal da Transparencia.

## Available Tools

### search_contracts
Search federal government contracts by organization.
- **orgao_code**: SIAPE organization code (default: 36000 - Ministry of Health)
- **page**: Page number (default: 1)
- **page_size**: Results per page (max: 500)

### search_servidores
Search federal public servants by name.
- **nome**: Name to search (required)
- **page**: Page number (default: 1)
- **page_size**: Results per page (max: 500)

### get_remuneracao
Get salary data for a public servant.
- **cpf**: CPF number (required, 11 digits)
- **mes_ano**: Month/Year in MM/YYYY format (default: last month)

### search_convenios
Search government agreements by state.
- **uf**: State code like MG, SP, RJ (default: MG)
- **page**: Page number (default: 1)
- **page_size**: Results per page (max: 500)

### search_ceis
Search sanctioned companies.
- **cnpj**: Company CNPJ (optional)
- **page**: Page number (default: 1)
- **page_size**: Results per page (max: 500)

### list_orgaos
List known organization codes for use with other tools.

## Known Organization Codes
%s

## Data Source
All data comes from Portal da Transparencia: https://api.portaldatransparencia.gov.br

## Rate Limits
The API has rate limits. Use pagination to avoid hitting limits.
`, string(orgaosJSON))
}
