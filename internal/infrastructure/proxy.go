package infrastructure

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/lucianocarvalho/labelify/internal/usecase"
)

type contextKey string

const queryKey contextKey = "originalQuery"

type Proxy struct {
	proxy   *httputil.ReverseProxy
	hydrate *usecase.HydrateUseCase
}

func NewProxy(targetURL string, hydrate *usecase.HydrateUseCase) (*Proxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		proxy:   httputil.NewSingleHostReverseProxy(target),
		hydrate: hydrate,
	}, nil
}

// ServeHTTP implementa a interface http.Handler
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	log.Printf("Query recebida no proxy: '%s'", query)

	ctx := context.WithValue(r.Context(), queryKey, query)
	r = r.WithContext(ctx)

	p.proxy.ServeHTTP(w, r)
}

// SetupModifyResponse configura a modificação da resposta
func (p *Proxy) SetupModifyResponse() {
	p.proxy.ModifyResponse = func(resp *http.Response) error {
		var body []byte
		var err error

		if resp.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(resp.Body)
			if err != nil {
				log.Printf("Erro ao criar gzip reader: %v", err)
				return err
			}
			defer reader.Close()
			body, err = io.ReadAll(reader)
			if err != nil {
				log.Printf("Erro ao ler body gzip: %v", err)
				return err
			}
			resp.Header.Del("Content-Encoding")
		} else {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Erro ao ler body: %v", err)
				return err
			}
		}
		resp.Body.Close()

		query := ""
		if val := resp.Request.Context().Value(queryKey); val != nil {
			query = val.(string)
		}
		log.Printf("Query recuperada do contexto: '%s'", query)

		newBody, err := p.hydrate.Execute(body, query)
		if err != nil {
			log.Printf("Erro ao executar hidratação: %v", err)
			newBody = body
		}

		resp.Body = io.NopCloser(bytes.NewReader(newBody))
		resp.ContentLength = int64(len(newBody))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))

		return nil
	}
}
