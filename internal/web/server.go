package web

import (
	"context"
	"encoding/json"
	"net/http"

	"linsky-backend/internal/service"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
)

// Start inicia el servidor HTTP en addr y devuelve *http.Server. El servidor se ejecuta en goroutine.
func Start(ctx context.Context, svc *service.URLService, addr string) *http.Server {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	// Middlewares: CORS simple
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// Endpoints API
	api.HandleFunc("/shorten", shortenHandler(svc)).Methods("POST", "OPTIONS")
	api.HandleFunc("/url/{alias}", getURLHandler(svc)).Methods("GET", "OPTIONS")
	api.HandleFunc("/urls", listHandler(svc)).Methods("GET", "OPTIONS")
	api.HandleFunc("/url/{alias}", deleteHandler(svc)).Methods("DELETE", "OPTIONS")

	// Redirect short path: /{alias}
	r.HandleFunc("/{alias}", redirectHandler(svc)).Methods("GET")

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		color.New(color.FgGreen).Printf("[web] Servidor escuchando en %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			color.New(color.FgRed).Printf("[web] Error servidor: %v\n", err)
		}
	}()

	return srv
}

func jsonResp(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func shortenHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Alias string `json:"alias"`
			URL   string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonResp(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		u, err := svc.Shorten(r.Context(), req.Alias, req.URL)
		if err != nil {
			jsonResp(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		jsonResp(w, http.StatusCreated, u)
	}
}

func getURLHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		alias := vars["alias"]
		u, err := svc.Resolve(r.Context(), alias)
		if err != nil {
			jsonResp(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		jsonResp(w, http.StatusOK, u)
	}
}

func listHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urls, err := svc.List(r.Context())
		if err != nil {
			jsonResp(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		jsonResp(w, http.StatusOK, urls)
	}
}

func deleteHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		alias := vars["alias"]
		if err := svc.Delete(r.Context(), alias); err != nil {
			jsonResp(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		jsonResp(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func redirectHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		alias := vars["alias"]
		u, err := svc.Resolve(r.Context(), alias)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		// Redirigir (no JSON)
		http.Redirect(w, r, u.OriginalURL, http.StatusFound)
	}
}
