package db

import (
	"context"
	"fmt"

	"linsky-backend/internal/config"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(cfg *config.Config) (*pgxpool.Pool, error) {
    pool, err := pgxpool.New(context.Background(), cfg.DSN())
    if err != nil {
        return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
    }

    // Verificar conexión
    if err := pool.Ping(context.Background()); err != nil {
        return nil, fmt.Errorf("no se pudo hacer ping a la BD: %w", err)
    }

    color.New(color.FgGreen).Println("✅ Conectado a Supabase PostgreSQL")
    return pool, nil
}
