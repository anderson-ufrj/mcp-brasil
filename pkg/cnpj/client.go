// Package cnpj provides a client for Minha Receita API (CNPJ lookup).
package cnpj

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	BaseURL        = "https://minhareceita.org"
	DefaultTimeout = 30 * time.Second
)

// Client represents the Minha Receita API client.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Minha Receita client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// CNPJData represents company data from Minha Receita.
type CNPJData struct {
	CNPJ                       string                   `json:"cnpj"`
	RazaoSocial                string                   `json:"razao_social"`
	NomeFantasia               string                   `json:"nome_fantasia,omitempty"`
	SituacaoCadastral          int                      `json:"situacao_cadastral"`
	DescricaoSituacaoCadastral string                   `json:"descricao_situacao_cadastral,omitempty"`
	DataSituacaoCadastral      string                   `json:"data_situacao_cadastral,omitempty"`
	AtividadePrincipal         map[string]interface{}   `json:"atividade_principal,omitempty"`
	AtividadesSecundarias      []map[string]interface{} `json:"atividades_secundarias,omitempty"`
	NaturezaJuridica           string                   `json:"natureza_juridica,omitempty"`
	Logradouro                 string                   `json:"logradouro,omitempty"`
	Numero                     string                   `json:"numero,omitempty"`
	Complemento                string                   `json:"complemento,omitempty"`
	Bairro                     string                   `json:"bairro,omitempty"`
	Municipio                  string                   `json:"municipio,omitempty"`
	UF                         string                   `json:"uf,omitempty"`
	CEP                        string                   `json:"cep,omitempty"`
	Email                      string                   `json:"email,omitempty"`
	Telefone                   string                   `json:"telefone,omitempty"`
	DataAbertura               string                   `json:"data_abertura,omitempty"`
	CapitalSocial              float64                  `json:"capital_social,omitempty"`
	QSA                        []Partner                `json:"qsa,omitempty"`
	Source                     string                   `json:"source"`
}

// Partner represents a company partner (QSA - Quadro Societario).
type Partner struct {
	Nome                 string `json:"nome_socio,omitempty"`
	CPFRepresentante     string `json:"cpf_representante_legal,omitempty"`
	NomeRepresentante    string `json:"nome_representante_legal,omitempty"`
	QualificacaoSocio    string `json:"qualificacao_socio,omitempty"`
	DataEntradaSociedade string `json:"data_entrada_sociedade,omitempty"`
}

// formatCNPJ formats a CNPJ string to the API format (XX.XXX.XXX/XXXX-XX).
func formatCNPJ(cnpj string) (string, error) {
	// Remove all non-digits
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, cnpj)

	if len(digits) != 14 {
		return "", fmt.Errorf("invalid CNPJ: must have 14 digits, got %d", len(digits))
	}

	// Format: XX.XXX.XXX/XXXX-XX
	return fmt.Sprintf("%s.%s.%s/%s-%s",
		digits[0:2], digits[2:5], digits[5:8], digits[8:12], digits[12:14]), nil
}

// GetCNPJ retrieves company data by CNPJ.
func (c *Client) GetCNPJ(ctx context.Context, cnpj string) (*CNPJData, error) {
	formattedCNPJ, err := formatCNPJ(cnpj)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", BaseURL, formattedCNPJ)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("CNPJ not found: %s", formattedCNPJ)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data CNPJData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	data.Source = "minhareceita_api"
	return &data, nil
}
