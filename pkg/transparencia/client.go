// Package transparencia provides a client for the Brazilian Portal da Transparencia API.
package transparencia

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL        = "https://api.portaldatransparencia.gov.br/api-de-dados"
	DefaultTimeout = 30 * time.Second
)

// Known organization codes (SIAPE)
var KnownOrgaos = map[string]string{
	"36000": "Ministério da Saúde",
	"26000": "Ministério da Educação",
	"25000": "Ministério da Economia",
	"30000": "Ministério da Justiça",
	"52000": "Ministério da Defesa",
	"35000": "Ministério das Relações Exteriores",
	"44000": "Ministério do Meio Ambiente",
}

// Client represents the Portal da Transparencia API client.
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// NewClient creates a new Portal da Transparencia client.
func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		apiKey:     apiKey,
		baseURL:    BaseURL,
	}
}

// doRequest performs an HTTP request to the API.
func (c *Client) doRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	if len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MCP-Brasil/1.0 (Go)")
	if c.apiKey != "" {
		req.Header.Set("chave-api-dados", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Contract represents a government contract.
type Contract struct {
	ID                 int64   `json:"id"`
	Numero             string  `json:"numero"`
	Objeto             string  `json:"objeto"`
	NumeroProcesso     string  `json:"numeroProcesso"`
	FundamentoLegal    string  `json:"fundamentoLegal"`
	DataAssinatura     string  `json:"dataAssinatura"`
	DataVigenciaInicio string  `json:"dataVigenciaInicio"`
	DataVigenciaFim    string  `json:"dataVigenciaFim"`
	ValorInicial       float64 `json:"valorInicial"`
	Situacao           string  `json:"situacao"`
	ModalidadeCompra   string  `json:"modalidadeCompra"`
	CodigoOrgao        string  `json:"codigoOrgao"`
	NomeOrgao          string  `json:"nomeOrgao"`
	CNPJFornecedor     string  `json:"cnpjFornecedor"`
	NomeFornecedor     string  `json:"nomeFornecedor"`
}

// ContractsResponse represents the API response for contracts.
type ContractsResponse struct {
	Contracts []Contract `json:"contratos"`
	Total     int        `json:"total"`
	Page      int        `json:"pagina"`
	PageSize  int        `json:"tamanhoPagina"`
	OrgaoCode string     `json:"orgaoConsultado"`
	OrgaoName string     `json:"orgaoNome"`
	Source    string     `json:"source"`
}

// SearchContracts searches for government contracts.
func (c *Client) SearchContracts(ctx context.Context, orgaoCode string, page, pageSize int) (*ContractsResponse, error) {
	if orgaoCode == "" {
		orgaoCode = "36000" // Default: Ministerio da Saude
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 100
	}

	params := url.Values{}
	params.Set("codigoOrgao", orgaoCode)
	params.Set("pagina", fmt.Sprintf("%d", page))
	params.Set("tamanhoPagina", fmt.Sprintf("%d", pageSize))

	body, err := c.doRequest(ctx, "/contratos", params)
	if err != nil {
		return nil, err
	}

	var contracts []Contract
	if err := json.Unmarshal(body, &contracts); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	orgaoName := KnownOrgaos[orgaoCode]
	if orgaoName == "" {
		orgaoName = "Orgao Desconhecido"
	}

	return &ContractsResponse{
		Contracts: contracts,
		Total:     len(contracts),
		Page:      page,
		PageSize:  pageSize,
		OrgaoCode: orgaoCode,
		OrgaoName: orgaoName,
		Source:    "portal_transparencia_api",
	}, nil
}

// Servidor represents a public servant.
type Servidor struct {
	ID               int64   `json:"id"`
	CPF              string  `json:"cpf"`
	Nome             string  `json:"nome"`
	Matricula        string  `json:"matricula"`
	CodigoOrgao      string  `json:"codigoOrgaoLotacao"`
	NomeOrgao        string  `json:"nomeOrgaoLotacao"`
	CodigoUorg       string  `json:"codigoUorgLotacao"`
	NomeUorg         string  `json:"nomeUorgLotacao"`
	TipoVinculo      string  `json:"tipoVinculo"`
	SituacaoVinculo  string  `json:"situacaoVinculo"`
	DataIngressoCarg string  `json:"dataIngressoCargo"`
}

// ServidoresResponse represents the API response for public servants.
type ServidoresResponse struct {
	Servidores []Servidor `json:"servidores"`
	Total      int        `json:"total"`
	Page       int        `json:"pagina"`
	PageSize   int        `json:"tamanhoPagina"`
	Source     string     `json:"source"`
}

// SearchServidores searches for public servants by name.
func (c *Client) SearchServidores(ctx context.Context, nome string, page, pageSize int) (*ServidoresResponse, error) {
	if nome == "" {
		return nil, fmt.Errorf("nome is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 100
	}

	params := url.Values{}
	params.Set("nome", nome)
	params.Set("pagina", fmt.Sprintf("%d", page))
	params.Set("tamanhoPagina", fmt.Sprintf("%d", pageSize))

	body, err := c.doRequest(ctx, "/servidores", params)
	if err != nil {
		return nil, err
	}

	var servidores []Servidor
	if err := json.Unmarshal(body, &servidores); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &ServidoresResponse{
		Servidores: servidores,
		Total:      len(servidores),
		Page:       page,
		PageSize:   pageSize,
		Source:     "portal_transparencia_api",
	}, nil
}

// Remuneracao represents a public servant's salary.
type Remuneracao struct {
	MesAno                 string  `json:"mesAno"`
	RemuneracaoBasicaBruta float64 `json:"remuneracaoBasicaBruta"`
	AbateGratificacao      float64 `json:"abateGratificacao"`
	GratificacaoNatalina   float64 `json:"gratificacaoNatalina"`
	AbateTeto              float64 `json:"abateTeto"`
	RendimentoLiquido      float64 `json:"rendimentoLiquido"`
}

// RemuneracaoResponse represents the API response for salary data.
type RemuneracaoResponse struct {
	CPF         string        `json:"cpf"`
	Remuneracao []Remuneracao `json:"remuneracao"`
	MesAno      string        `json:"mesAno"`
	Source      string        `json:"source"`
}

// GetServidorRemuneracao gets salary data for a public servant by CPF.
func (c *Client) GetServidorRemuneracao(ctx context.Context, cpf, mesAno string) (*RemuneracaoResponse, error) {
	if cpf == "" {
		return nil, fmt.Errorf("cpf is required")
	}
	if mesAno == "" {
		// Default to last month
		lastMonth := time.Now().AddDate(0, -1, 0)
		mesAno = lastMonth.Format("01/2006")
	}

	params := url.Values{}
	params.Set("mesAno", mesAno)

	endpoint := fmt.Sprintf("/servidores/%s/remuneracao", cpf)
	body, err := c.doRequest(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}

	var remuneracoes []Remuneracao
	if err := json.Unmarshal(body, &remuneracoes); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &RemuneracaoResponse{
		CPF:         cpf,
		Remuneracao: remuneracoes,
		MesAno:      mesAno,
		Source:      "portal_transparencia_api",
	}, nil
}

// Convenio represents a government agreement/covenant.
type Convenio struct {
	Numero          string  `json:"numero"`
	Objeto          string  `json:"objeto"`
	SituacaoConveni string  `json:"situacaoConvenio"`
	ValorLiberado   float64 `json:"valorLiberado"`
	ValorConvenio   float64 `json:"valorConvenio"`
	UF              string  `json:"uf"`
	Municipio       string  `json:"municipio"`
	OrgaoSuperior   string  `json:"orgaoSuperior"`
	DataInicio      string  `json:"dataInicioVigencia"`
	DataFim         string  `json:"dataFimVigencia"`
}

// ConveniosResponse represents the API response for agreements.
type ConveniosResponse struct {
	Convenios []Convenio `json:"convenios"`
	Total     int        `json:"total"`
	Page      int        `json:"pagina"`
	PageSize  int        `json:"tamanhoPagina"`
	UF        string     `json:"uf"`
	Source    string     `json:"source"`
}

// SearchConvenios searches for government agreements by state.
func (c *Client) SearchConvenios(ctx context.Context, uf string, page, pageSize int) (*ConveniosResponse, error) {
	if uf == "" {
		uf = "MG" // Default: Minas Gerais
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 100
	}

	params := url.Values{}
	params.Set("uf", uf)
	params.Set("pagina", fmt.Sprintf("%d", page))
	params.Set("tamanhoPagina", fmt.Sprintf("%d", pageSize))

	body, err := c.doRequest(ctx, "/convenios", params)
	if err != nil {
		return nil, err
	}

	var convenios []Convenio
	if err := json.Unmarshal(body, &convenios); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &ConveniosResponse{
		Convenios: convenios,
		Total:     len(convenios),
		Page:      page,
		PageSize:  pageSize,
		UF:        uf,
		Source:    "portal_transparencia_api",
	}, nil
}

// CEIS represents a company in the sanctions list.
type CEIS struct {
	CNPJ            string `json:"cnpjSancionado"`
	RazaoSocial     string `json:"razaoSocialSancionado"`
	NomeFantasia    string `json:"nomeFantasia"`
	TipoSancao      string `json:"tipoSancao"`
	DataInicioSanca string `json:"dataInicioSancao"`
	DataFimSancao   string `json:"dataFimSancao"`
	OrgaoSancionado string `json:"orgaoSancionador"`
}

// CEISResponse represents the API response for sanctions.
type CEISResponse struct {
	Empresas []CEIS `json:"empresas"`
	Total    int    `json:"total"`
	Page     int    `json:"pagina"`
	PageSize int    `json:"tamanhoPagina"`
	Source   string `json:"source"`
}

// SearchCEIS searches for sanctioned companies.
func (c *Client) SearchCEIS(ctx context.Context, cnpj string, page, pageSize int) (*CEISResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 100
	}

	params := url.Values{}
	if cnpj != "" {
		params.Set("cnpj", cnpj)
	}
	params.Set("pagina", fmt.Sprintf("%d", page))
	params.Set("tamanhoPagina", fmt.Sprintf("%d", pageSize))

	body, err := c.doRequest(ctx, "/ceis", params)
	if err != nil {
		return nil, err
	}

	var empresas []CEIS
	if err := json.Unmarshal(body, &empresas); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &CEISResponse{
		Empresas: empresas,
		Total:    len(empresas),
		Page:     page,
		PageSize: pageSize,
		Source:   "portal_transparencia_api",
	}, nil
}

// ListOrgaos returns the list of known organization codes.
func (c *Client) ListOrgaos() []map[string]string {
	result := make([]map[string]string, 0, len(KnownOrgaos))
	for code, name := range KnownOrgaos {
		result = append(result, map[string]string{
			"codigo": code,
			"nome":   name,
		})
	}
	return result
}
