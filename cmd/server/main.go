// MCP Brasil - Model Context Protocol server for Brazilian government data.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anderson-ufrj/mcp-brasil/pkg/bcb"
	"github.com/anderson-ufrj/mcp-brasil/pkg/cnpj"
	"github.com/anderson-ufrj/mcp-brasil/pkg/ibge"
	"github.com/anderson-ufrj/mcp-brasil/pkg/pncp"
	"github.com/anderson-ufrj/mcp-brasil/pkg/transparencia"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	transparenciaClient *transparencia.Client
	ibgeClient          *ibge.Client
	cnpjClient          *cnpj.Client
	bcbClient           *bcb.Client
	pncpClient          *pncp.Client
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("TRANSPARENCY_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Warning: TRANSPARENCY_API_KEY not set, some features may not work")
	}

	// Initialize clients
	transparenciaClient = transparencia.NewClient(apiKey)
	ibgeClient = ibge.NewClient()
	cnpjClient = cnpj.NewClient()
	bcbClient = bcb.NewClient()
	pncpClient = pncp.NewClient()

	// Create MCP server
	s := server.NewMCPServer(
		"MCP Brasil",
		"2.0.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(true, false),
	)

	// Register all tools
	registerTransparenciaTools(s)
	registerIBGETools(s)
	registerCNPJTools(s)
	registerBCBTools(s)
	registerPNCPTools(s)

	// Register resources
	registerResources(s)

	// Run server over stdio
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// ==================== PORTAL DA TRANSPARENCIA ====================

func registerTransparenciaTools(s *server.MCPServer) {
	// search_contracts
	s.AddTool(mcp.NewTool("search_contracts",
		mcp.WithDescription("Search government contracts from Portal da Transparencia"),
		mcp.WithString("orgao_code", mcp.Description("Organization SIAPE code (e.g. 36000 for Ministry of Health)")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("page_size", mcp.Description("Results per page (max 500)")),
	), handleSearchContracts)

	// search_servidores
	s.AddTool(mcp.NewTool("search_servidores",
		mcp.WithDescription("Search federal public servants by name"),
		mcp.WithString("nome", mcp.Required(), mcp.Description("Name of the public servant")),
		mcp.WithNumber("page", mcp.Description("Page number")),
		mcp.WithNumber("page_size", mcp.Description("Results per page")),
	), handleSearchServidores)

	// get_remuneracao
	s.AddTool(mcp.NewTool("get_remuneracao",
		mcp.WithDescription("Get salary data for a public servant by CPF"),
		mcp.WithString("cpf", mcp.Required(), mcp.Description("CPF (11 digits)")),
		mcp.WithString("mes_ano", mcp.Description("Month/Year MM/YYYY format")),
	), handleGetRemuneracao)

	// search_convenios
	s.AddTool(mcp.NewTool("search_convenios",
		mcp.WithDescription("Search federal government agreements by state"),
		mcp.WithString("uf", mcp.Description("State code (e.g. MG, SP, RJ)")),
		mcp.WithNumber("page", mcp.Description("Page number")),
		mcp.WithNumber("page_size", mcp.Description("Results per page")),
	), handleSearchConvenios)

	// search_ceis
	s.AddTool(mcp.NewTool("search_ceis",
		mcp.WithDescription("Search sanctioned companies in CEIS"),
		mcp.WithString("cnpj", mcp.Description("Company CNPJ (optional)")),
		mcp.WithNumber("page", mcp.Description("Page number")),
		mcp.WithNumber("page_size", mcp.Description("Results per page")),
	), handleSearchCEIS)

	// list_orgaos
	s.AddTool(mcp.NewTool("list_orgaos",
		mcp.WithDescription("List known government organization codes (SIAPE)"),
	), handleListOrgaos)
}

// ==================== IBGE ====================

func registerIBGETools(s *server.MCPServer) {
	// ibge_states
	s.AddTool(mcp.NewTool("ibge_states",
		mcp.WithDescription("List all Brazilian states with their codes and regions"),
	), handleIBGEStates)

	// ibge_municipalities
	s.AddTool(mcp.NewTool("ibge_municipalities",
		mcp.WithDescription("List municipalities, optionally filtered by state"),
		mcp.WithString("state_id", mcp.Description("State ID (e.g. 33 for RJ, 35 for SP). Leave empty for all.")),
	), handleIBGEMunicipalities)

	// ibge_population
	s.AddTool(mcp.NewTool("ibge_population",
		mcp.WithDescription("Get population data for Brazil or a specific location"),
		mcp.WithString("location_id", mcp.Description("Municipality IBGE code (optional)")),
	), handleIBGEPopulation)
}

// ==================== CNPJ (Minha Receita) ====================

func registerCNPJTools(s *server.MCPServer) {
	// lookup_cnpj
	s.AddTool(mcp.NewTool("lookup_cnpj",
		mcp.WithDescription("Look up company data by CNPJ. Returns registration info, address, partners (QSA), and economic activity."),
		mcp.WithString("cnpj", mcp.Required(), mcp.Description("CNPJ (14 digits, with or without formatting)")),
	), handleLookupCNPJ)
}

// ==================== BANCO CENTRAL ====================

func registerBCBTools(s *server.MCPServer) {
	// bcb_selic
	s.AddTool(mcp.NewTool("bcb_selic",
		mcp.WithDescription("Get SELIC interest rate data from Banco Central"),
		mcp.WithNumber("last_n", mcp.Description("Number of data points to retrieve (default 30)")),
	), handleBCBSelic)

	// bcb_ipca
	s.AddTool(mcp.NewTool("bcb_ipca",
		mcp.WithDescription("Get IPCA (inflation index) data from Banco Central"),
		mcp.WithNumber("last_n", mcp.Description("Number of months to retrieve (default 12)")),
	), handleBCBIPCA)

	// bcb_exchange_rate
	s.AddTool(mcp.NewTool("bcb_exchange_rate",
		mcp.WithDescription("Get exchange rate for a currency (USD, EUR, etc.)"),
		mcp.WithString("currency", mcp.Description("Currency code (default USD)")),
		mcp.WithString("date", mcp.Description("Date in MM-DD-YYYY format (default today)")),
	), handleBCBExchangeRate)

	// bcb_indicator
	s.AddTool(mcp.NewTool("bcb_indicator",
		mcp.WithDescription("Get any economic indicator: selic, selic_monthly, ipca, igpm, cdi"),
		mcp.WithString("indicator", mcp.Required(), mcp.Description("Indicator name")),
		mcp.WithNumber("last_n", mcp.Description("Number of data points")),
	), handleBCBIndicator)
}

// ==================== PNCP ====================

func registerPNCPTools(s *server.MCPServer) {
	// pncp_contracts
	s.AddTool(mcp.NewTool("pncp_contracts",
		mcp.WithDescription("Search public procurement contracts from PNCP (Portal Nacional de Contratacoes Publicas)"),
		mcp.WithString("start_date", mcp.Required(), mcp.Description("Start date YYYYMMDD format")),
		mcp.WithString("end_date", mcp.Required(), mcp.Description("End date YYYYMMDD format")),
		mcp.WithString("state", mcp.Description("State code (e.g. SP, RJ)")),
		mcp.WithNumber("modality", mcp.Description("Procurement modality code (default 6 = pregao eletronico)")),
		mcp.WithNumber("page", mcp.Description("Page number")),
	), handlePNCPContracts)

	// pncp_modalities
	s.AddTool(mcp.NewTool("pncp_modalities",
		mcp.WithDescription("List available procurement modality codes for PNCP queries"),
	), handlePNCPModalities)
}

// ==================== RESOURCES ====================

func registerResources(s *server.MCPServer) {
	docResource := mcp.NewResource(
		"docs://api-reference",
		"API Reference",
		mcp.WithResourceDescription("Documentation for all Brazilian government APIs"),
		mcp.WithMIMEType("text/markdown"),
	)
	s.AddResource(docResource, handleDocResource)
}

// ==================== HANDLERS: Portal da Transparencia ====================

func handleSearchContracts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	orgaoCode, _ := request.GetArguments()["orgao_code"].(string)
	page := getIntArg(request, "page", 1)
	pageSize := getIntArg(request, "page_size", 100)

	result, err := transparenciaClient.SearchContracts(ctx, orgaoCode, page, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleSearchServidores(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nome, _ := request.RequireString("nome")
	page := getIntArg(request, "page", 1)
	pageSize := getIntArg(request, "page_size", 100)

	result, err := transparenciaClient.SearchServidores(ctx, nome, page, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleGetRemuneracao(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cpf, _ := request.RequireString("cpf")
	mesAno, _ := request.GetArguments()["mes_ano"].(string)

	result, err := transparenciaClient.GetServidorRemuneracao(ctx, cpf, mesAno)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleSearchConvenios(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uf, _ := request.GetArguments()["uf"].(string)
	page := getIntArg(request, "page", 1)
	pageSize := getIntArg(request, "page_size", 100)

	result, err := transparenciaClient.SearchConvenios(ctx, uf, page, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleSearchCEIS(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cnpj, _ := request.GetArguments()["cnpj"].(string)
	page := getIntArg(request, "page", 1)
	pageSize := getIntArg(request, "page_size", 100)

	result, err := transparenciaClient.SearchCEIS(ctx, cnpj, page, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleListOrgaos(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return toJSONResult(transparenciaClient.ListOrgaos())
}

// ==================== HANDLERS: IBGE ====================

func handleIBGEStates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := ibgeClient.GetStates(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleIBGEMunicipalities(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stateID, _ := request.GetArguments()["state_id"].(string)

	result, err := ibgeClient.GetMunicipalities(ctx, stateID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleIBGEPopulation(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	locationID, _ := request.GetArguments()["location_id"].(string)

	result, err := ibgeClient.GetPopulation(ctx, locationID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

// ==================== HANDLERS: CNPJ ====================

func handleLookupCNPJ(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cnpjNum, err := request.RequireString("cnpj")
	if err != nil {
		return mcp.NewToolResultError("Parameter 'cnpj' is required"), nil
	}

	result, err := cnpjClient.GetCNPJ(ctx, cnpjNum)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

// ==================== HANDLERS: BCB ====================

func handleBCBSelic(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lastN := getIntArg(request, "last_n", 30)

	result, err := bcbClient.GetSELIC(ctx, lastN)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleBCBIPCA(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lastN := getIntArg(request, "last_n", 12)

	result, err := bcbClient.GetIPCA(ctx, lastN)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleBCBExchangeRate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	currency, _ := request.GetArguments()["currency"].(string)
	date, _ := request.GetArguments()["date"].(string)

	result, err := bcbClient.GetExchangeRate(ctx, currency, date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handleBCBIndicator(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indicator, err := request.RequireString("indicator")
	if err != nil {
		return mcp.NewToolResultError("Parameter 'indicator' is required"), nil
	}
	lastN := getIntArg(request, "last_n", 30)

	result, err := bcbClient.GetIndicator(ctx, indicator, lastN)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

// ==================== HANDLERS: PNCP ====================

func handlePNCPContracts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startDate, _ := request.RequireString("start_date")
	endDate, _ := request.RequireString("end_date")
	state, _ := request.GetArguments()["state"].(string)
	modality := getIntArg(request, "modality", 6)
	page := getIntArg(request, "page", 1)

	result, err := pncpClient.SearchContracts(ctx, startDate, endDate, modality, state, page, 50)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}
	return toJSONResult(result)
}

func handlePNCPModalities(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return toJSONResult(pncpClient.ListModalities())
}

// ==================== HANDLERS: Resources ====================

func handleDocResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "docs://api-reference",
			MIMEType: "text/markdown",
			Text:     getAPIDocumentation(),
		},
	}, nil
}

// ==================== HELPERS ====================

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
	return `# MCP Brasil - API Reference v2.0

## Overview
This MCP server provides access to multiple Brazilian government data sources.

## Available Tools

### Portal da Transparencia (Federal Government)
| Tool | Description |
|------|-------------|
| search_contracts | Search federal government contracts |
| search_servidores | Search public servants by name |
| get_remuneracao | Get salary by CPF |
| search_convenios | Search agreements by state |
| search_ceis | Search sanctioned companies |
| list_orgaos | List organization codes |

### IBGE (Statistics)
| Tool | Description |
|------|-------------|
| ibge_states | List all Brazilian states |
| ibge_municipalities | List municipalities (filter by state) |
| ibge_population | Get population data |

### CNPJ Lookup (Minha Receita)
| Tool | Description |
|------|-------------|
| lookup_cnpj | Get company data by CNPJ |

### Banco Central (Economic Data)
| Tool | Description |
|------|-------------|
| bcb_selic | Get SELIC interest rate |
| bcb_ipca | Get IPCA inflation index |
| bcb_exchange_rate | Get exchange rates |
| bcb_indicator | Get any indicator (selic, ipca, igpm, cdi) |

### PNCP (Public Procurement)
| Tool | Description |
|------|-------------|
| pncp_contracts | Search procurement contracts |
| pncp_modalities | List procurement modalities |

## Data Sources
- Portal da Transparencia: https://api.portaldatransparencia.gov.br
- IBGE: https://servicodados.ibge.gov.br
- Minha Receita: https://minhareceita.org
- Banco Central: https://api.bcb.gov.br
- PNCP: https://pncp.gov.br
`
}
