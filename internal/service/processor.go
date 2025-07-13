package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/config"
)

type ProcessorService struct {
	transport         *http.Transport
	client            *http.Client
	defaultURL        string
	fallbackURL       string
	defaultTimeout    time.Duration
	fallbackTimeout   time.Duration
	processorAPIToken string
	mu                sync.RWMutex
}

const (
	ProcessorDefault  = "default"
	ProcessorFallback = "fallback"

	processorDefaultTimeout  = 500 * time.Millisecond
	processorFallbackTimeout = 500 * time.Millisecond
)

func NewProcessorService(cfg *config.Config) *ProcessorService {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	cl := &http.Client{Transport: tr}

	return &ProcessorService{
		transport:         tr,
		client:            cl,
		defaultURL:        cfg.DefaultProcessorURL,
		fallbackURL:       cfg.FallBackProcessorURL,
		defaultTimeout:    processorDefaultTimeout,
		fallbackTimeout:   processorFallbackTimeout,
		processorAPIToken: cfg.ProcessorAPIToken,
	}
}

func (p *ProcessorService) SetDefaultTimeout(t int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.defaultTimeout = time.Duration(t) * time.Millisecond
}

func (p *ProcessorService) SetFallbackTimeout(t int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.fallbackTimeout = time.Duration(t) * time.Millisecond
}

func (p *ProcessorService) MakeRequestDefault(method, path string, body io.Reader) (io.ReadCloser, int, error) {
	return p.MakeRequest(ProcessorDefault, method, path, body)
}

func (p *ProcessorService) MakeRequestFallback(method, path string, body io.Reader) (io.ReadCloser, int, error) {
	return p.MakeRequest(ProcessorFallback, method, path, body)
}

func (p *ProcessorService) MakeRequest(processor, method, path string, body io.Reader) (io.ReadCloser, int, error) {
	var reqPath string
	var timeout time.Duration

	switch processor {
	case ProcessorDefault:
		reqPath = fmt.Sprintf("%s%s", p.defaultURL, path)
		timeout = p.defaultTimeout
	case ProcessorFallback:
		reqPath = fmt.Sprintf("%s%s", p.fallbackURL, path)
		timeout = p.fallbackTimeout
	default:
		return nil, 0, errors.New("invalid processor")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		reqPath,
		body,
	)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Rinha-Token", p.processorAPIToken)

	res, err := p.client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	return res.Body, res.StatusCode, nil
}
