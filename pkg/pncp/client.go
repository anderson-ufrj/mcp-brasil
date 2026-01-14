// Package pncp provides a client for PNCP (Portal Nacional de Contratacoes Publicas) API.
package pncp

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
	BaseURL        = "https://pncp.gov.br/api/consulta/v1"
	DefaultTimeout = 30 * time.Second
)

// Procurement modality codes.
var Modalities = map[string]int{
	"pregao_eletronico":       6,
	"concorrencia_eletronica": 1,
	"concorrencia":            2,
	"concurso":                3,
	"leilao_eletronico":       4,
	"leilao":                  5,
	"dialogo_competitivo":     7,
	"credenciamento":          8,
}

// Client represents the PNCP API client.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new PNCP client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// ContractPublication represents a contract publication from PNCP.
type ContractPublication struct {
	SequencialCompra          int                    `json:"sequencialCompra,omitempty"`
	NumeroCompra              string                 `json:"numeroCompra,omitempty"`
	AnoCompra                 int                    `json:"anoCompra,omitempty"`
	OrgaoEntidade             map[string]interface{} `json:"orgaoEntidade,omitempty"`
	ModalidadeID              int                    `json:"modalidadeId,omitempty"`
	ModalidadeNome            string                 `json:"modalidadeNome,omitempty"`
	SituacaoCompraID          int                    `json:"situacaoCompraId,omitempty"`
	SituacaoCompraNome        string                 `json:"situacaoCompraNome,omitempty"`
	NumeroControlePNCP        string                 `json:"numeroControlePNCP,omitempty"`
	DataPublicacaoPncp        string                 `json:"dataPublicacaoPncp,omitempty"`
	DataAberturaProposta      string                 `json:"dataAberturaProposta,omitempty"`
	DataEncerramentoProposta  string                 `json:"dataEncerramentoProposta,omitempty"`
	ObjetoCompra              string                 `json:"objetoCompra,omitempty"`
	ValorTotalEstimado        float64                `json:"valorTotalEstimado,omitempty"`
	ValorTotalHomologado      float64                `json:"valorTotalHomologado,omitempty"`
}

// ContractsResponse represents the response for contracts query.
type ContractsResponse struct {
	Contracts []ContractPublication `json:"contracts"`
	Total     int                   `json:"total"`
	Page      int                   `json:"page"`
	PageSize  int                   `json:"page_size"`
	Source    string                `json:"source"`
}

// PriceRegistration represents a price registration record.
type PriceRegistration struct {
	NumeroControlePNCP  string                 `json:"numeroControlePNCP,omitempty"`
	OrgaoEntidade       map[string]interface{} `json:"orgaoEntidade,omitempty"`
	NumeroAta           string                 `json:"numeroAta,omitempty"`
	AnoAta              int                    `json:"anoAta,omitempty"`
	DataPublicacaoPncp  string                 `json:"dataPublicacaoPncp,omitempty"`
	DataVigenciaInicio  string                 `json:"dataVigenciaInicio,omitempty"`
	DataVigenciaFim     string                 `json:"dataVigenciaFim,omitempty"`
	ObjetoAta           string                 `json:"objetoAta,omitempty"`
	ValorTotalEstimado  float64                `json:"valorTotalEstimado,omitempty"`
}

// PriceRegistrationsResponse represents the response for price registrations query.
type PriceRegistrationsResponse struct {
	Registrations []PriceRegistration `json:"registrations"`
	Total         int                 `json:"total"`
	Page          int                 `json:"page"`
	Source        string              `json:"source"`
}

func (c *Client) doRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s", BaseURL, endpoint)
	if len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

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

// SearchContracts searches for contract publications.
func (c *Client) SearchContracts(ctx context.Context, startDate, endDate string, modalityCode int, state string, page, pageSize int) (*ContractsResponse, error) {
	if pageSize < 10 {
		pageSize = 10
	} else if pageSize > 500 {
		pageSize = 500
	}
	if page < 1 {
		page = 1
	}
	if modalityCode == 0 {
		modalityCode = 6 // Default: pregao eletronico
	}

	params := url.Values{}
	params.Set("dataInicial", startDate)
	params.Set("dataFinal", endDate)
	params.Set("codigoModalidadeContratacao", fmt.Sprintf("%d", modalityCode))
	params.Set("tamanhoPagina", fmt.Sprintf("%d", pageSize))
	params.Set("pagina", fmt.Sprintf("%d", page))

	if state != "" {
		params.Set("uf", state)
	}

	body, err := c.doRequest(ctx, "/contratacoes/publicacao", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data          []ContractPublication `json:"data"`
		TotalRegistros int                  `json:"totalRegistros"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &ContractsResponse{
		Contracts: result.Data,
		Total:     result.TotalRegistros,
		Page:      page,
		PageSize:  pageSize,
		Source:    "pncp_api",
	}, nil
}

// SearchPriceRegistrations searches for price registration records.
func (c *Client) SearchPriceRegistrations(ctx context.Context, state string, page, pageSize int) (*PriceRegistrationsResponse, error) {
	if pageSize < 10 {
		pageSize = 10
	} else if pageSize > 500 {
		pageSize = 500
	}
	if page < 1 {
		page = 1
	}

	params := url.Values{}
	params.Set("tamanhoPagina", fmt.Sprintf("%d", pageSize))
	params.Set("pagina", fmt.Sprintf("%d", page))

	if state != "" {
		params.Set("uf", state)
	}

	body, err := c.doRequest(ctx, "/atas-registro-preco", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []PriceRegistration `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &PriceRegistrationsResponse{
		Registrations: result.Data,
		Total:         len(result.Data),
		Page:          page,
		Source:        "pncp_api",
	}, nil
}

// ListModalities returns available procurement modalities.
func (c *Client) ListModalities() map[string]int {
	return Modalities
}
