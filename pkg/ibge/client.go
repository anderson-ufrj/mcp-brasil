// Package ibge provides a client for the IBGE (Brazilian Institute of Geography and Statistics) API.
package ibge

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	LocalidadesURL = "https://servicodados.ibge.gov.br/api/v1/localidades"
	AgregadosURL   = "https://servicodados.ibge.gov.br/api/v3/agregados"
	DefaultTimeout = 30 * time.Second
)

// Client represents the IBGE API client.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new IBGE client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// State represents a Brazilian state.
type State struct {
	ID     int    `json:"id"`
	Sigla  string `json:"sigla"`
	Nome   string `json:"nome"`
	Regiao Region `json:"regiao"`
}

// Region represents a Brazilian region.
type Region struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

// Municipality represents a Brazilian municipality.
type Municipality struct {
	ID          int    `json:"id"`
	Nome        string `json:"nome"`
	Microrregiao struct {
		ID   int    `json:"id"`
		Nome string `json:"nome"`
	} `json:"microrregiao"`
}

// StatesResponse represents the response for states query.
type StatesResponse struct {
	States []State `json:"states"`
	Total  int     `json:"total"`
	Source string  `json:"source"`
}

// MunicipalitiesResponse represents the response for municipalities query.
type MunicipalitiesResponse struct {
	Municipalities []Municipality `json:"municipalities"`
	Total          int            `json:"total"`
	StateID        string         `json:"state_id,omitempty"`
	Source         string         `json:"source"`
}

// PopulationData represents population data.
type PopulationData struct {
	Location   string `json:"location"`
	Year       string `json:"year"`
	Population string `json:"population"`
}

// PopulationResponse represents the response for population query.
type PopulationResponse struct {
	Data   []PopulationData `json:"data"`
	Source string           `json:"source"`
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

// GetStates returns all Brazilian states.
func (c *Client) GetStates(ctx context.Context) (*StatesResponse, error) {
	url := fmt.Sprintf("%s/estados?orderBy=nome", LocalidadesURL)

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var states []State
	if err := json.Unmarshal(body, &states); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &StatesResponse{
		States: states,
		Total:  len(states),
		Source: "ibge_api",
	}, nil
}

// GetMunicipalities returns municipalities, optionally filtered by state.
func (c *Client) GetMunicipalities(ctx context.Context, stateID string) (*MunicipalitiesResponse, error) {
	var url string
	if stateID != "" {
		url = fmt.Sprintf("%s/estados/%s/municipios?orderBy=nome", LocalidadesURL, stateID)
	} else {
		url = fmt.Sprintf("%s/municipios?orderBy=nome", LocalidadesURL)
	}

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var municipalities []Municipality
	if err := json.Unmarshal(body, &municipalities); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &MunicipalitiesResponse{
		Municipalities: municipalities,
		Total:          len(municipalities),
		StateID:        stateID,
		Source:         "ibge_api",
	}, nil
}

// GetPopulation returns population data for a location.
func (c *Client) GetPopulation(ctx context.Context, locationID string) (*PopulationResponse, error) {
	// Population estimate (agregado 6579, variable 9324)
	var url string
	if locationID != "" {
		url = fmt.Sprintf("%s/6579/periodos/-6/variaveis/9324?localidades=N6[%s]", AgregadosURL, locationID)
	} else {
		url = fmt.Sprintf("%s/6579/periodos/-6/variaveis/9324?localidades=N1[all]", AgregadosURL)
	}

	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var data []PopulationData
	if len(result) > 0 {
		if resultados, ok := result[0]["resultados"].([]interface{}); ok && len(resultados) > 0 {
			if series, ok := resultados[0].(map[string]interface{})["series"].([]interface{}); ok {
				for _, s := range series {
					serie := s.(map[string]interface{})
					localidade := serie["localidade"].(map[string]interface{})
					if serieData, ok := serie["serie"].(map[string]interface{}); ok {
						for year, pop := range serieData {
							data = append(data, PopulationData{
								Location:   localidade["nome"].(string),
								Year:       year,
								Population: fmt.Sprintf("%v", pop),
							})
						}
					}
				}
			}
		}
	}

	return &PopulationResponse{
		Data:   data,
		Source: "ibge_api",
	}, nil
}
