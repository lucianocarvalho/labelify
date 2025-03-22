package infrastructure

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/lucianocarvalho/labelify/internal/usecase"
)

type Proxy struct {
	proxy      *httputil.ReverseProxy
	enrichment *usecase.EnrichmentUseCase
}

func NewProxy(targetURL string, enrichment *usecase.EnrichmentUseCase) (*Proxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		proxy:      httputil.NewSingleHostReverseProxy(target),
		enrichment: enrichment,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the path is not /api/v1/query or /api/v1/query_range, serve the request as is
	if r.URL.Path != "/api/v1/query" && r.URL.Path != "/api/v1/query_range" {
		p.proxy.ServeHTTP(w, r)
		return
	}

	query := r.URL.Query().Get("query")
	p.proxy.ModifyResponse = func(resp *http.Response) error {
		var body []byte
		var err error

		// Prometheus uses a gzip compression for the response body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer reader.Close()
			body, err = io.ReadAll(reader)
			if err != nil {
				return err
			}
			resp.Header.Del("Content-Encoding")
		} else {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
		}
		resp.Body.Close()

		newBody, err := p.enrichment.Execute(body, query)
		if err != nil {
			log.Printf("Erro ao executar hidratação: %v", err)
			newBody = body
		}

		resp.Body = io.NopCloser(bytes.NewReader(newBody))
		resp.ContentLength = int64(len(newBody))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
		return nil
	}

	p.proxy.ServeHTTP(w, r)
}
