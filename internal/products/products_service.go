package products

import (
	"context"

	repo "github.com/johnbkh98/ecom/internal/adapters/postgresql/sqlc"
)

type Service interface {
	ListProducts(ctx context.Context) ([]repo.Product, error)
	FindProductById(ctx context.Context, product_id int64) (repo.Product, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) ListProducts(ctx context.Context) ([]repo.Product, error) {
	return s.repo.ListProducts(ctx)
}

func (s *svc) FindProductById(ctx context.Context, product_id int64) (repo.Product, error) {
	return s.repo.FindProductById(ctx, product_id)
}
