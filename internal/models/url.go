package models

import "time"

type ShortURL struct {
    ID          int64     `json:"id"`
    Alias       string    `json:"alias"`        // El alias personalizado (ej: "mi-blog")
    OriginalURL string    `json:"original_url"` // La URL larga original
    Clicks      int64     `json:"clicks"`       // Contador de visitas
    CreatedAt   time.Time `json:"created_at"`
}
