package repository

import (
	"context"
	"fmt"

	"linsky-backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepository struct {
    db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) *URLRepository {
    return &URLRepository{db: db}
}

// Crear la tabla si no existe
func (r *URLRepository) Migrate(ctx context.Context) error {
    query := `
    CREATE TABLE IF NOT EXISTS short_urls (
        id          BIGSERIAL PRIMARY KEY,
        alias       VARCHAR(50) UNIQUE NOT NULL,
        original_url TEXT NOT NULL,
        clicks      BIGINT DEFAULT 0,
        created_at  TIMESTAMPTZ DEFAULT NOW()
    );`
    _, err := r.db.Exec(ctx, query)
    if err != nil {
        return fmt.Errorf("error al migrar: %w", err)
    }
    return nil
}

// Guardar una nueva URL corta
func (r *URLRepository) Create(ctx context.Context, alias, originalURL string) (*models.ShortURL, error) {
    query := `
    INSERT INTO short_urls (alias, original_url)
    VALUES ($1, $2)
    RETURNING id, alias, original_url, clicks, created_at`

    row := r.db.QueryRow(ctx, query, alias, originalURL)

    var u models.ShortURL
    err := row.Scan(&u.ID, &u.Alias, &u.OriginalURL, &u.Clicks, &u.CreatedAt)
    if err != nil {
        return nil, fmt.Errorf("error al crear URL: %w", err)
    }
    return &u, nil
}

// Buscar por alias
func (r *URLRepository) FindByAlias(ctx context.Context, alias string) (*models.ShortURL, error) {
    query := `SELECT id, alias, original_url, clicks, created_at FROM short_urls WHERE alias = $1`

    row := r.db.QueryRow(ctx, query, alias)

    var u models.ShortURL
    err := row.Scan(&u.ID, &u.Alias, &u.OriginalURL, &u.Clicks, &u.CreatedAt)
    if err != nil {
        return nil, fmt.Errorf("alias '%s' no encontrado", alias)
    }
    return &u, nil
}

// Incrementar clicks
func (r *URLRepository) IncrementClicks(ctx context.Context, alias string) error {
    _, err := r.db.Exec(ctx, `UPDATE short_urls SET clicks = clicks + 1 WHERE alias = $1`, alias)
    return err
}

// Listar todas
func (r *URLRepository) ListAll(ctx context.Context) ([]models.ShortURL, error) {
    rows, err := r.db.Query(ctx, `SELECT id, alias, original_url, clicks, created_at FROM short_urls ORDER BY created_at DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var urls []models.ShortURL
    for rows.Next() {
        var u models.ShortURL
        if err := rows.Scan(&u.ID, &u.Alias, &u.OriginalURL, &u.Clicks, &u.CreatedAt); err != nil {
            return nil, err
        }
        urls = append(urls, u)
    }
    return urls, nil
}

// Eliminar por alias
func (r *URLRepository) Delete(ctx context.Context, alias string) error {
    result, err := r.db.Exec(ctx, `DELETE FROM short_urls WHERE alias = $1`, alias)
    if err != nil {
        return err
    }
    if result.RowsAffected() == 0 {
        return fmt.Errorf("alias '%s' no existe", alias)
    }
    return nil
}
