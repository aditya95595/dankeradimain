package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net"
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

// loginAttempts tracks recent failed dashboard logins per client IP so
// brute-force attacks are throttled instead of running unopposed.
var loginAttempts = struct {
	sync.Mutex
	hits map[string][]time.Time
}{hits: make(map[string][]time.Time)}

const (
	// loginMaxAttempts is the number of failed logins allowed per
	// loginWindow for a single client IP before further attempts are
	// rejected outright.
	loginMaxAttempts = 5
	loginWindow      = time.Minute
	// loginFailDelay is slept after every failed login so automated
	// guessing runs slowly even below the rate limit.
	loginFailDelay = 750 * time.Millisecond
	// sessionLifetime bounds how long one dashboard login stays valid.
	sessionLifetime = 24 * time.Hour
)

func startWebServer(dmgService *DmgService, assets embed.FS) {
	initDashboardAuth()
	go sweepExpiredSessions()
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
			view["accountIndex"] = accountIndex(dmgService.GetConfig(), in.AccountCfg.Token)
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
			AccountIndex    int    `json:"accountIndex"`
			ReadyState      string `json:"readyState"`
			BreakUpdateTime string `json:"breakUpdateTime"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body); err != nil {
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
			var body struct {
				Restarting bool `json:"restarting"`
			}
			_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body)
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
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&account); err != nil ||
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
		var body struct {
			Status types.OnlineStatus `json:"status"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body); err != nil {
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
	if err := http.ListenAndServe(addr, securityHeaders(mux)); err != nil {
		slog.Error("Web server error", "error", err)
	}
}

// securityHeaders applies baseline hardening headers to every response.
// The dashboard is a control panel: it must never be framed, must never be
// MIME-sniffed, and must not leak its URL as a referrer.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

// sweepExpiredSessions periodically purges expired dashboard sessions so
// they do not accumulate in memory over long uptimes.
func sweepExpiredSessions() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		dashboardAuth.Lock()
		for token, expiry := range dashboardAuth.sessions {
			if now.After(expiry) {
				delete(dashboardAuth.sessions, token)
			}
		}
		dashboardAuth.Unlock()
	}
}

// clientIP returns the best-effort client IP, preferring X-Forwarded-For
// (set by hosting proxies) over the direct remote address.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// loginAllowed reports whether this client IP may attempt a login right
// now, pruning attempts that fell outside the window.
func loginAllowed(ip string) bool {
	now := time.Now()
	loginAttempts.Lock()
	defer loginAttempts.Unlock()
	hits := loginAttempts.hits[ip]
	fresh := hits[:0]
	for _, t := range hits {
		if now.Sub(t) < loginWindow {
			fresh = append(fresh, t)
		}
	}
	loginAttempts.hits[ip] = fresh
	return len(fresh) < loginMaxAttempts
}

func recordLoginFailure(ip string) {
	loginAttempts.Lock()
	loginAttempts.hits[ip] = append(loginAttempts.hits[ip], time.Now())
	loginAttempts.Unlock()
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
		// Fail closed: the random password is never displayed, so the
		// dashboard cannot be logged into until DASHBOARD_PASSWORD is set.
		slog.Warn("DASHBOARD_PASSWORD is not set; dashboard login is DISABLED until you set it as a secret on your host.")
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
	ip := clientIP(r)
	if !loginAllowed(ip) {
		slog.Warn("dashboard login rate limit exceeded", "ip", ip)
		http.Error(w, "too many attempts, try again later", http.StatusTooManyRequests)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	dashboardAuth.RLock()
	password := dashboardAuth.password
	dashboardAuth.RUnlock()
	// Constant-time compare: response time must not leak password bytes.
	ok := body.Password != "" && subtle.ConstantTimeCompare([]byte(body.Password), []byte(password)) == 1
	if !ok {
		recordLoginFailure(ip)
		slog.Warn("failed dashboard login", "ip", ip)
		time.Sleep(loginFailDelay)
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
	dashboardAuth.sessions[token] = time.Now().Add(sessionLifetime)
	dashboardAuth.Unlock()
	// Secure requires HTTPS; the dashboard must be served over TLS or the
	// browser will refuse to store/send this cookie and login will fail.
	http.SetCookie(w, &http.Cookie{
		Name: "dmg_session", Value: token, Path: "/", HttpOnly: true,
		Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: int(sessionLifetime.Seconds()),
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
		Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: -1,
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

func accountIndex(cfg *config.Config, token string) int {
	if cfg == nil {
		return -1
	}
	for i, account := range cfg.Accounts {
		if account.Token == token {
			return i
		}
	}
	return -1
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
