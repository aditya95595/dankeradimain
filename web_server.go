package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/autocord-org/dmg/config"
	"github.com/autocord-org/dmg/discord/types"
)

func startWebServer(dmgService *DmgService, assets embed.FS) {
	mux := http.NewServeMux()

	// ── SSE events stream ────────────────────────────────────────────
	mux.HandleFunc("/api/events", sseHandler)

	// ── Config ───────────────────────────────────────────────────────
	mux.HandleFunc("/api/config", withCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, dmgService.GetConfig())
		case http.MethodPost:
			var cfg config.Config
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := dmgService.UpdateConfig(&cfg); err != nil {
				writeJSONError(w, err.Error())
				return
			}
			writeJSON(w, map[string]bool{"success": true})
		}
	}))

	// ── Discord status ────────────────────────────────────────────────
	mux.HandleFunc("/api/discord-status", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Status types.OnlineStatus `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		dmgService.UpdateDiscordStatus(body.Status)
		writeJSON(w, map[string]bool{"success": true})
	}))

	// ── Check for updates ────────────────────────────────────────────
	mux.HandleFunc("/api/check-updates", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		hasUpdate := dmgService.CheckForUpdates()
		writeJSON(w, map[string]bool{"hasUpdate": hasUpdate})
	}))

	// ── Self-update ──────────────────────────────────────────────────
	mux.HandleFunc("/api/update", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		go dmgService.Update()
		writeJSON(w, map[string]bool{"success": true})
	}))

	// ── Instance token update ────────────────────────────────────────
	mux.HandleFunc("/api/instances/token", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			OldToken string `json:"oldToken"`
			NewToken string `json:"newToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		dmgService.UpdateInstanceToken(body.OldToken, body.NewToken)
		writeJSON(w, map[string]bool{"success": true})
	}))

	// ── Restart all instances ────────────────────────────────────────
	mux.HandleFunc("/api/instances/restart", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		go dmgService.RestartInstances()
		writeJSON(w, map[string]bool{"success": true})
	}))

	// ── Start instance ────────────────────────────────────────────────
	mux.HandleFunc("/api/instances/start", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Account         config.AccountsConfig `json:"account"`
			ReadyState      string                `json:"readyState"`
			BreakUpdateTime string                `json:"breakUpdateTime"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		t, err := time.Parse(time.RFC3339, body.BreakUpdateTime)
		if err != nil {
			t = time.Now()
		}
		go dmgService.StartInstance(body.Account, body.ReadyState, t)
		writeJSON(w, map[string]bool{"success": true})
	}))

	// ── Per-instance routes: /api/instances/{token} ──────────────────
	mux.HandleFunc("/api/instances/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/instances/")
		parts := strings.SplitN(rest, "/", 2)
		token := parts[0]

		if len(parts) == 1 {
			if r.Method != http.MethodDelete {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var body struct {
				Restarting bool `json:"restarting"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			dmgService.RemoveInstance(token, body.Restarting)
			writeJSON(w, map[string]bool{"success": true})
			return
		}

		if parts[1] == "restart" && r.Method == http.MethodPost {
			view := dmgService.RestartInstance(token)
			writeJSON(w, view)
			return
		}

		http.NotFound(w, r)
	}))

	// ── Health check (UptimeRobot / host health probes) ───────────────
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// ── Static frontend assets ────────────────────────────────────────
	distFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		slog.Error("Failed to sub frontend/dist", "error", err)
	} else {
		mux.Handle("/", http.FileServer(http.FS(distFS)))
	}

	// Wispbyte and most hosts inject PORT. Default to 5000 for local/Replit.
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	addr := "0.0.0.0:" + port
	slog.Info("Web dashboard running", "url", "http://"+addr)
	slog.Info("Health check", "path", "/health")
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("Web server error", "error", err)
	}
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
