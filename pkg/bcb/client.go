// Package bcb provides a client for the Banco Central do Brasil API.
package bcb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	SGSURL         = "https://api.bcb.gov.br/dados/serie/bcdata.sgs"
	OlindaURL      = "https://olinda.bcb.gov.br/olinda/servico"
	DefaultTimeout = 30 * time.Second
)

// Series codes for economic indicators.
var SeriesCodes = map[string]int{
	"selic":         11,   // SELIC daily
	"selic_monthly": 4390, // SELIC accumulated monthly
	"ipca":          433,  // IPCA monthly
	"igpm":          189,  // IGP-M monthly
	"cdi":           12,   // CDI daily
}

// Client represents the BCB API client.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new BCB client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// DataPoint represents a single data point from BCB.
type DataPoint struct {
	Date  string `json:"data"`
	Value string `json:"valor"`
}

// IndicatorResponse represents the response for indicator queries.
type IndicatorResponse struct {
	Indicator string      `json:"indicator"`
	Data      []DataPoint `json:"data"`
	Total     int         `json:"total"`
	Source    string      `json:"source"`
}

// ExchangeRate represents an exchange rate data point.
type ExchangeRate struct {
	DateTime     string  `json:"dataHoraCotacao"`
	BuyRate      float64 `json:"cotacaoCompra"`
	SellRate     float64 `json:"cotacaoVenda"`
	BulletinType string  `json:"tipoBoletim"`
}

// ExchangeRateResponse represents the response for exchange rate queries.
type ExchangeRateResponse struct {
	Currency string         `json:"currency"`
	Date     string         `json:"date"`
	Rates    []ExchangeRate `json:"rates"`
	Source   string         `json:"source"`
}

// PIXStats represents PIX statistics.
type PIXStats struct {
	TotalTransactions int64   `json:"total_transactions,omitempty"`
	TotalValue        float64 `json:"total_value,omitempty"`
	Data              interface{} `json:"data,omitempty"`
}

// PIXResponse represents the response for PIX statistics.
type PIXResponse struct {
	Stats  PIXStats `json:"stats"`
	Source string   `json:"source"`
}

func (c *Client) doRequest(ctx context.Context, url string) ([]byte, error) {
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetIndicator retrieves economic indicator data.
func (c *Client) GetIndicator(ctx context.Context, indicator string, lastN int) (*IndicatorResponse, error) {
	seriesCode, ok := SeriesCodes[indicator]
	if !ok {
		return nil, fmt.Errorf("unknown indicator: %s. Available: selic, selic_monthly, ipca, igpm, cdi", indicator)
	}

	if lastN <= 0 {
		lastN = 30 // Default to last 30 values
	}

	url := fmt.Sprintf("%s.%d/dados/ultimos/%d?formato=json", SGSURL, seriesCode, lastN)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var data []DataPoint
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &IndicatorResponse{
		Indicator: indicator,
		Data:      data,
		Total:     len(data),
		Source:    "bcb_api",
	}, nil
}

// GetSELIC retrieves SELIC rate data.
func (c *Client) GetSELIC(ctx context.Context, lastN int) (*IndicatorResponse, error) {
	return c.GetIndicator(ctx, "selic", lastN)
}

// GetIPCA retrieves IPCA (inflation) data.
func (c *Client) GetIPCA(ctx context.Context, lastN int) (*IndicatorResponse, error) {
	return c.GetIndicator(ctx, "ipca", lastN)
}

// GetExchangeRate retrieves exchange rate for a currency.
func (c *Client) GetExchangeRate(ctx context.Context, currency, date string) (*ExchangeRateResponse, error) {
	if currency == "" {
		currency = "USD"
	}
	if date == "" {
		date = time.Now().Format("01-02-2006")
	}

	url := fmt.Sprintf("%s/PTAX/versao/v1/odata/CotacaoMoedaDia(moeda=@moeda,dataCotacao=@dataCotacao)?@moeda='%s'&@dataCotacao='%s'&$format=json",
		OlindaURL, currency, date)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var result struct {
		Value []ExchangeRate `json:"value"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &ExchangeRateResponse{
		Currency: currency,
		Date:     date,
		Rates:    result.Value,
		Source:   "bcb_api",
	}, nil
}

// GetPIXStats retrieves PIX statistics.
func (c *Client) GetPIXStats(ctx context.Context) (*PIXResponse, error) {
	url := fmt.Sprintf("%s/Pix_DadosAbertos/versao/v1/odata/EstatisticasTransacoesPix(Database=@Database)?@Database='202401'&$format=json", OlindaURL)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &PIXResponse{
		Stats: PIXStats{
			Data: result,
		},
		Source: "bcb_api",
	}, nil
}
