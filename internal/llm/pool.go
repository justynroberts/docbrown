package llm

import (
	"context"
	"fmt"
	"sync"
)

// Pool manages concurrent LLM operations
type Pool struct {
	provider      Provider
	semaphore     chan struct{}
	maxConcurrent int
	mu            sync.Mutex
	totalCost     float64
	totalTokens   int
}

// NewPool creates a new LLM pool
func NewPool(provider Provider, maxConcurrent int) *Pool {
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}

	return &Pool{
		provider:      provider,
		semaphore:     make(chan struct{}, maxConcurrent),
		maxConcurrent: maxConcurrent,
	}
}

// Execute executes a function with concurrency control
func (p *Pool) Execute(ctx context.Context, fn func() error) error {
	select {
	case p.semaphore <- struct{}{}:
		defer func() { <-p.semaphore }()
		return fn()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Analyze performs analysis with concurrency control
func (p *Pool) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResult, error) {
	var result *AnalysisResult
	var err error

	execErr := p.Execute(ctx, func() error {
		result, err = p.provider.Analyze(ctx, req)
		return err
	})

	if execErr != nil {
		return nil, execErr
	}

	return result, err
}

// Generate performs generation with concurrency control
func (p *Pool) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	var result string
	var err error

	execErr := p.Execute(ctx, func() error {
		result, err = p.provider.Generate(ctx, req)
		return err
	})

	if execErr != nil {
		return "", execErr
	}

	return result, err
}

// GenerateParallel generates documentation for multiple components in parallel
func (p *Pool) GenerateParallel(ctx context.Context, requests []GenerateRequest) ([]string, error) {
	results := make([]string, len(requests))
	errors := make([]error, len(requests))

	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)

		go func(idx int, request GenerateRequest) {
			defer wg.Done()

			result, err := p.Generate(ctx, request)
			results[idx] = result
			errors[idx] = err
		}(i, req)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return results, fmt.Errorf("generation failed for request %d: %w", i, err)
		}
	}

	return results, nil
}

// GetProvider returns the underlying provider
func (p *Pool) GetProvider() Provider {
	return p.provider
}

// TrackCost tracks the cost of an operation
func (p *Pool) TrackCost(tokens int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.totalTokens += tokens
	p.totalCost += p.provider.EstimateCost(tokens)
}

// GetTotalCost returns the total cost
func (p *Pool) GetTotalCost() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.totalCost
}

// GetTotalTokens returns the total tokens used
func (p *Pool) GetTotalTokens() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.totalTokens
}
