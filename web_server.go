package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/autocord-org/dmg/config"
	"github.com/autocord-org/dmg/discord/types"
)

var dashboardAuth = struct {
	sync.RWMutex
	password string
	sessions map[string]time.Time
}{sessions: make(map[string]time.Time)}

func startWebServer(dmgService *DmgService, assets embed.FS) {
	initDashboardAuth()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/logout", requireAuth(logoutHandler))
	mux.HandleFunc("/api/session", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]bool{"authenticated": true})
	}))
	mux.HandleFunc("/api/events", requireAuth(sseHandler))

	mux.HandleFunc("/api/config", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, publicConfig(dmgService.GetConfig()))
		case http.MethodPost:
			var incoming config.Config
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&incoming); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mergeProtectedConfig(dmgService.GetConfig(), &incoming)
			if err := dmgService.UpdateConfig(&incoming); err != nil {
				writeJSONError(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, map[string]bool{"success": true})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/instances", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		views := make([]map[string]interface{}, 0, len(dmgService.instances))
		for _, in := range dmgService.instances {
			if in == nil {
				continue
			}
			raw, _ := json.Marshal(in.GetView())
			var view map[string]interface{}
			_ = json.Unmarshal(raw, &view)
			view["instanceId"] = instanceID(in.AccountCfg.Token)
			if account, ok := view["accountCfg"].(map[string]interface{}); ok {
				account["token"] = ""
			}
			views = append(views, view)
		}
		writeJSON(w, views)
	}))

	mux.HandleFunc("/api/instances/start", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			AccountIndex int
			ReadyState string
			BreakUpdateTime string
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		cfg := dmgService.GetConfig()
		if body.AccountIndex < 0 || body.AccountIndex >= len(cfg.Accounts) {
			http.Error(w, "invalid account index", http.StatusBadRequest)
			return
		}
		t, err := time.Parse(time.RFC3339, body.BreakUpdateTime)
		if err != nil {
			t = time.Now()
		}
		ready := body.ReadyState
		if ready == "" {
			ready = "ready"
		}
		go dmgService.StartInstance(cfg.Accounts[body.AccountIndex], ready, t)
		writeJSON(w, map[string]bool{"success": true})
	}))

	mux.HandleFunc("/api/instances/restart", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		go dmgService.RestartInstances()
		writeJSON(w, map[string]bool{"success": true})
	}))

	mux.HandleFunc("/api/instances/", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/instances/")
		parts := strings.SplitN(rest, "/", 2)
		token := dmgService.tokenForInstanceID(parts[0])
		if token == "" {
			http.NotFound(w, r)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodDelete {
			var body struct { Restarting bool }
			_ = json.NewDecoder(r.Body).Decode(&body)
			dmgService.RemoveInstance(token, body.Restarting)
			writeJSON(w, map[string]bool{"success": true})
			return
		}
		if len(parts) == 2 && parts[1] == "restart" && r.Method == http.MethodPost {
			writeJSON(w, dmgService.RestartInstance(token))
			return
		}
		http.NotFound(w, r)
	}))

	mux.HandleFunc("/api/accounts", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var account config.AccountsConfig
		if err := json.NewDecoder(r.Body).Decode(&account); err != nil ||
			strings.TrimSpace(account.Token) == "" || strings.TrimSpace(account.ChannelID) == "" {
			http.Error(w, "token and channelID are required", http.StatusBadRequest)
			return
		}
		cfg := *dmgService.GetConfig()
		account.State = true
		cfg.Accounts = append(cfg.Accounts, account)
		if err := dmgService.UpdateConfig(&cfg); err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		dmgService.StartInstance(account, "ready", time.Now())
		writeJSON(w, map[string]bool{"success": true})
	}))

	mux.HandleFunc("/api/discord-status", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct { Status types.OnlineStatus }
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		dmgService.UpdateDiscordStatus(body.Status)
		writeJSON(w, map[string]bool{"success": true})
	}))

	mux.HandleFunc("/api/check-updates", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]bool{"hasUpdate": dmgService.CheckForUpdates()})
	}))

	mux.HandleFunc("/api/update", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		go dmgService.Update()
		writeJSON(w, map[string]bool{"success": true})
	}))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"status\":\"ok\"}"))
	})

	distFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		slog.Error("Failed to sub frontend/dist", "error", err)
	} else {
		mux.Handle("/", http.FileServer(http.FS(distFS)))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	addr := "0.0.0.0:" + port
	slog.Info("Web dashboard running", "url", "http://"+addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("Web server error", "error", err)
	}
}

func initDashboardAuth() {
	password := strings.TrimSpace(os.Getenv("DASHBOARD_PASSWORD"))
	if password == "" {
		buf := make([]byte, 18)
		if _, err := rand.Read(buf); err != nil {
			password = "temporary-dashboard-password"
		} else {
			password = hex.EncodeToString(buf)
		}
		slog.Warn("DASHBOARD_PASSWORD is not set; a temporary dashboard password was generated. Set DASHBOARD_PASSWORD in Wispbyte secrets.")
	}
	dashboardAuth.Lock()
	dashboardAuth.password = password
	dashboardAuth.Unlock()
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct { Password string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	dashboardAuth.RLock()
	ok := body.Password != "" && body.Password == dashboardAuth.password
	dashboardAuth.RUnlock()
	if !ok {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(buf)
	dashboardAuth.Lock()
	dashboardAuth.sessions[token] = time.Now().Add(24 * time.Hour)
	dashboardAuth.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name: "dmg_session", Value: token, Path: "/", HttpOnly: true,
		Secure: true, SameSite: http.SameSiteStrict, MaxAge: 86400,
	})
	writeJSON(w, map[string]bool{"success": true})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("dmg_session"); err == nil {
		dashboardAuth.Lock()
		delete(dashboardAuth.sessions, c.Value)
		dashboardAuth.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name: "dmg_session", Value: "", Path: "/", HttpOnly: true,
		Secure: true, SameSite: http.SameSiteStrict, MaxAge: -1,
	})
	writeJSON(w, map[string]bool{"success": true})
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("dmg_session")
		if err != nil || !validSession(c.Value) {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func validSession(token string) bool {
	dashboardAuth.Lock()
	defer dashboardAuth.Unlock()
	expiry, ok := dashboardAuth.sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(dashboardAuth.sessions, token)
		return false
	}
	return true
}

func publicConfig(src *config.Config) config.Config {
	out := *src
	out.ApiKey = ""
	out.Accounts = make([]config.AccountsConfig, len(src.Accounts))
	for i, account := range src.Accounts {
		out.Accounts[i] = account
		out.Accounts[i].Token = ""
	}
	return out
}

func mergeProtectedConfig(current, incoming *config.Config) {
	if incoming.ApiKey == "" {
		incoming.ApiKey = current.ApiKey
	}
	for i := range incoming.Accounts {
		if incoming.Accounts[i].Token == "" && i < len(current.Accounts) {
			incoming.Accounts[i].Token = current.Accounts[i].Token
		}
		if incoming.Accounts[i].ChannelID == "" && i < len(current.Accounts) {
			incoming.Accounts[i].ChannelID = current.Accounts[i].ChannelID
		}
	}
}

func instanceID(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])[:12]
}

func (d *DmgService) tokenForInstanceID(id string) string {
	for _, in := range d.instances {
		if in != nil && instanceID(in.AccountCfg.Token) == id {
			return in.AccountCfg.Token
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
