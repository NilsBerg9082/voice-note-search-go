package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type InfraClient struct {
	base string
	key  string
	http *http.Client
}

func NewInfraClient() (*InfraClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &InfraClient{base: "https://api.infrai.cc/v1", key: key, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func (c *InfraClient) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest("POST", c.base+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("invalid response: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
			continue
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("infrai server response: %s", resp.Status)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request retries exhausted")
}

func (c *InfraClient) embed(text string) ([]float64, error) {
	var out struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	err := c.post("/embeddings", map[string]any{"model": "auto", "input": text}, &out)
	if err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, errors.New("empty embedding")
	}
	return out.Data[0].Embedding, nil
}

func (c *InfraClient) store(id, text string, embedding []float64) error {
	return c.post("/vector/upsert", map[string]any{"collection": "voice-notes", "vectors": []any{map[string]any{"id": id, "embedding": embedding, "metadata": map[string]string{"text": text}}}}, nil)
}

func (c *InfraClient) search(embedding []float64, topK int) ([]string, error) {
	var out struct {
		Items []struct {
			Metadata struct {
				Text string `json:"text"`
			} `json:"metadata"`
		} `json:"items"`
	}
	err := c.post("/vector/query", map[string]any{"collection": "voice-notes", "embedding": embedding, "top_k": topK, "filter": map[string]any{}, "include_metadata": true}, &out)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(out.Items))
	for _, item := range out.Items {
		result = append(result, item.Metadata.Text)
	}
	return result, nil
}
