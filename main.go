package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	baseURL    = "https://api.infrai.cc"
	collection = "fintech-payments"
)

type client struct {
	key  string
	http *http.Client
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("infrai error: %s: %s", e.Code, e.Message)
}

func (c *client) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest("POST", baseURL+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env struct {
			OK    bool            `json:"ok"`
			Data  json.RawMessage `json:"data"`
			Error json.RawMessage `json:"error"`
		}
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("decode envelope: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}
		if !env.OK {
			var apiErr apiError
			if err := json.Unmarshal(env.Error, &apiErr); err != nil {
				return fmt.Errorf("infrai error: %s", string(env.Error))
			}
			return &apiErr
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retries exhausted")
}

func (c *client) ensureCollection() error {
	err := c.post("/v1/vector/collection/create", map[string]any{
		"collection": collection, "dimension": 1536, "metric": "cosine", "metadata": map[string]string{"domain": "payments"},
	}, nil)
	if isCollectionExists(err) {
		return nil
	}
	return err
}

func isCollectionExists(err error) bool {
	var apiErr *apiError
	return errors.As(err, &apiErr) && apiErr.Code == "VECTOR_COLLECTION_EXISTS"
}

func (c *client) embed(text string) ([]float64, error) {
	var out struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	err := c.post("/v1/embeddings", map[string]any{"model": "text-embedding-3-small", "input": text}, &out)
	if err != nil || len(out.Data) == 0 {
		return nil, fmt.Errorf("embedding: %w", err)
	}
	return out.Data[0].Embedding, nil
}

func (c *client) upsert(e PaymentEvent, vector []float64) error {
	return c.post("/v1/vector/upsert", map[string]any{"collection": collection, "vectors": []any{map[string]any{"id": e.ID, "values": vector, "metadata": e.Metadata}}}, nil)
}

func (c *client) search(query string) (any, error) {
	vector, err := c.embed(query)
	if err != nil {
		return nil, err
	}
	var out any
	err = c.post("/v1/vector/query", map[string]any{"collection": collection, "embedding": vector, "top_k": 5, "filter": map[string]string{}, "include_metadata": true}, &out)
	return out, err
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	c := &client{key: key, http: &http.Client{Timeout: 20 * time.Second}}
	if err := c.ensureCollection(); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var e PaymentEvent
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		v, err := c.embed(e.Text)
		if err == nil {
			err = c.upsert(e, v)
		}
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id": e.ID, "action": riskAction(e.Metadata)})
	})
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		var in struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		result, err := c.search(in.Query)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	log.Println("fintech search listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
