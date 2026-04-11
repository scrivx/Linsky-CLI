package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"linsky-backend/internal/models"
	"linsky-backend/internal/repository"
)

var aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)

type URLService struct {
    repo    *repository.URLRepository
    baseURL string
}

func NewURLService(repo *repository.URLRepository, baseURL string) *URLService {
    return &URLService{repo: repo, baseURL: baseURL}
}

func (s *URLService) Shorten(ctx context.Context, alias, originalURL string) (*models.ShortURL, error) {
    // Validar alias
    if !aliasRegex.MatchString(alias) {
        return nil, fmt.Errorf("alias inválido: solo letras, números, guiones y _ (3-50 caracteres)")
    }

    // Validar URL
    if _, err := url.ParseRequestURI(originalURL); err != nil {
        return nil, fmt.Errorf("URL inválida: %s", originalURL)
    }

    return s.repo.Create(ctx, alias, originalURL)
}

func (s *URLService) Resolve(ctx context.Context, alias string) (*models.ShortURL, error) {
    u, err := s.repo.FindByAlias(ctx, alias)
    if err != nil {
        return nil, err
    }
    _ = s.repo.IncrementClicks(ctx, alias)
    return u, nil
}

func (s *URLService) List(ctx context.Context) ([]models.ShortURL, error) {
    return s.repo.ListAll(ctx)
}

func (s *URLService) Delete(ctx context.Context, alias string) error {
    return s.repo.Delete(ctx, alias)
}

func (s *URLService) ShortLink(alias string) string {
    return fmt.Sprintf("%s/%s", s.baseURL, alias)
}
