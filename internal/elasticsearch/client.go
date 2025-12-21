package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

type Client struct {
	es     *elasticsearch.Client
	logger *logrus.Logger
}

func NewClient(addresses []string, username, password string, log *logrus.Logger) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating elasticsearch client: %w", err)
	}

	// Test connection
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("error getting elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch returned error: %s", res.String())
	}

	log.Info("Successfully connected to Elasticsearch")

	return &Client{
		es:     es,
		logger: log,
	}, nil
}

func (c *Client) Index(ctx context.Context, index string, documentID string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       strings.NewReader(string(data)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from elasticsearch: %s", res.String())
	}

	return nil
}

func (c *Client) Search(ctx context.Context, index string, query map[string]interface{}) ([]byte, error) {
	var buf strings.Builder
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(index),
		c.es.Search.WithBody(strings.NewReader(buf.String())),
		c.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error performing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error response from elasticsearch: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) Delete(ctx context.Context, index string, documentID string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: documentID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("error deleting document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("error response from elasticsearch: %s", res.String())
	}

	return nil
}

func (c *Client) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	data, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("error marshaling mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(string(data)),
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 400 { // 400 means index already exists
		return fmt.Errorf("error response from elasticsearch: %s", res.String())
	}

	c.logger.WithField("index", index).Info("Elasticsearch index created or already exists")
	return nil
}
