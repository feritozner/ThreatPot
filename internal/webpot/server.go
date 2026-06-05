package webpot

import (
	"context"
	"embed"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"threatpot/internal/core"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

var activeServers = make(map[string]*http.Server)

const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
)

func StartServer(pot *core.Pot) error {
	if pot.IsRunning {
		return fmt.Errorf("pot is already running")
	}

	addr := fmt.Sprintf("%s:%d", pot.BindIP, pot.Port)
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if pot.Delay > 0 {
			randomDelay := rand.Intn(1001)
			time.Sleep(time.Duration(pot.Delay+randomDelay) * time.Millisecond)
		}

		dumpBytes, _ := httputil.DumpRequest(r, true)
		rawRequest := string(dumpBytes)

		pot.Mutex.Lock()
		pot.ReqCount++
		reqID := pot.ReqCount

		pot.ReqDumps = append(pot.ReqDumps, core.RequestDump{
			ID:      reqID,
			Time:    time.Now().Format("15:04:05"),
			Method:  r.Method,
			Path:    r.URL.Path,
			IP:      r.RemoteAddr,
			RawData: rawRequest,
		})
		pot.Mutex.Unlock()

		payloadPreview := ""
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			parts := strings.SplitN(rawRequest, "\r\n\r\n", 2)
			if len(parts) == 2 && len(parts[1]) > 0 {
				preview := parts[1]
				if len(preview) > 60 {
					preview = preview[:60] + "..."
				}
				payloadPreview = fmt.Sprintf("\n    └─ %s[PAYLOAD]%s %s", ColorRed, ColorReset, strings.TrimSpace(preview))
			}
		}

		logLine := fmt.Sprintf("[%sHIT #%d%s] %s%s%s %s%s%s | IP: %s%s%s | Host: %s%s%s | Agent: %s%s",
			ColorRed, reqID, ColorReset,
			ColorMagenta, r.Method, ColorReset,
			ColorYellow, r.URL.Path, ColorReset,
			ColorCyan, r.RemoteAddr, ColorReset,
			ColorBlue, r.Host, ColorReset,
			r.UserAgent(),
			payloadPreview)

		pot.AddLog(logLine)

		w.Header().Set("Server", pot.ServerHeader)

		pot.Mutex.Lock()
		htmlFile, routeExists := pot.Routes[r.URL.Path]
		pot.Mutex.Unlock()

		if routeExists {
			content, err := templateFS.ReadFile(htmlFile)
			if err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write(content)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("<html><body><h1>500 Internal Server Error</h1></body></html>"))
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html><body><h1>404 Not Found</h1></body></html>"))
	})

	server := &http.Server{Addr: addr, Handler: mux}
	activeServers[pot.Name] = server
	pot.IsRunning = true

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("\n[%sERROR%s] [%s] Server crashed: %v\n", ColorRed, ColorReset, pot.Name, err)
		}
		pot.IsRunning = false
	}()

	return nil
}

func StopServer(pot *core.Pot) error {
	if !pot.IsRunning {
		return fmt.Errorf("pot is not running")
	}

	if srv, ok := activeServers[pot.Name]; ok {
		srv.Shutdown(context.Background())
		delete(activeServers, pot.Name)
		pot.IsRunning = false
		return nil
	}

	return fmt.Errorf("server instance not found")
}
