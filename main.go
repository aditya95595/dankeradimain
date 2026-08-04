package main

import (
	"bufio"
	"context"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync"

	"github.com/autocord-org/dmg/utils"
	"github.com/grongor/panicwatch"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	redirectStderr()
}

func redirectStderr() {
	reader, writer, err := os.Pipe()
	if err != nil {
		slog.Error("Failed to create pipe for stderr redirection", "error", err)
		return
	}

	os.Stderr = writer

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "Overriding existing handler for signal") ||
				strings.Contains(line, `Failed to load module "appmenu-gtk-module"`) ||
				strings.Contains(line, "SetProcessDpiAwarenessContext failed 0") {
				continue
			}

			slog.Error(line)
		}

		if err := scanner.Err(); err != nil {
			slog.Error("Error reading from stderr pipe", "error", err)
		}
	}()
}

type CustomHandler struct {
	slog.Handler
}

func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := r.Time.Format("3:04PM")
	level := r.Level.String()[:3]
	msg := r.Message

	const (
		reset  = "\033[0m"
		red    = "\033[31m"
		green  = "\033[32m"
		yellow = "\033[33m"
		blue   = "\033[34m"
	)

	var color string
	switch r.Level {
	case slog.LevelDebug:
		color = blue
	case slog.LevelInfo:
		color = green
	case slog.LevelWarn:
		color = yellow
	case slog.LevelError:
		color = red
	default:
		color = reset
	}

	_, err := fmt.Fprintf(os.Stdout, "%s %s%s%s %s\n", timeStr, color, level, reset, msg)
	return err
}

func main() {
	customHandler := &CustomHandler{
		Handler: slog.NewTextHandler(os.Stdout, nil),
	}
	logger := slog.New(customHandler)
	slog.SetDefault(logger)

	dmgService := &DmgService{
		wg:      &sync.WaitGroup{},
		wsMutex: sync.Mutex{},
	}

	err := panicwatch.Start(panicwatch.Config{
		OnPanic: func(p panicwatch.Panic) {
			slog.Error("panic occurred",
				"message", p.Message,
				"stack", p.Stack)
		},
		OnWatcherDied: func(err error) {
			// In containerised environments (e.g. Replit) the watcher child
			// process can be reaped by the container runtime. Log the event
			// but do NOT exit — the main server must keep running.
			slog.Error("panicwatch watcher process died", "error", err)
		},
	})
	if err != nil {
		slog.Error("failed to start panicwatch: " + err.Error())
	}

	cli := flag.Bool("cli", false, "enable CLI mode")
	web := flag.Bool("web", false, "run as web server dashboard")
	flag.Parse()

	utils.SetCliMode(func() bool {
		return *cli || *web
	})

	if *web {
		// Web-server mode: serve the Svelte dashboard over HTTP and push
		// real-time events to connected browsers via SSE.
		utils.SetWebMode()
		utils.SetWebEventEmitter(func(eventName string, args ...interface{}) {
			eventHub.broadcast(eventName, args...)
		})
		dmgService.startup()
		startWebServer(dmgService, assets)
		return
	}

	if *cli {
		dmgService.startup()
		var wg sync.WaitGroup
		wg.Add(1)
		wg.Wait()
		return
	}

	// Neither -web nor -cli flag: default to web mode so the binary works
	// out-of-the-box on Replit (no display available).
	utils.SetWebMode()
	utils.SetWebEventEmitter(func(eventName string, args ...interface{}) {
		eventHub.broadcast(eventName, args...)
	})
	dmgService.startup()
	startWebServer(dmgService, assets)
}
