package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	open := flag.Bool("open", true, "Open browser automatically")
	flag.Parse()

	// Find dashboard directory (relative to this binary or workspace)
	dashboardDir := findDashboardDir()
	if dashboardDir == "" {
		log.Fatal("Could not find dashboard directory")
	}

	// Create file server
	fs := http.FileServer(http.Dir(dashboardDir))
	http.Handle("/", logRequest(fs))

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost%s", addr)

	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│  dependency-buster // Dashboard Server                  │")
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│  ▶ Serving: %-44s│\n", dashboardDir)
	fmt.Printf("│  ▶ URL:     %-44s│\n", url)
	fmt.Println("│  ▶ Press Ctrl+C to stop                                 │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Open browser
	if *open {
		openBrowser(url)
	}

	log.Fatal(http.ListenAndServe(addr, nil))
}

func findDashboardDir() string {
	// Try relative paths
	candidates := []string{
		"../dashboard",
		"./dashboard",
		"../../dpb-benchmark/dashboard",
		"../dpb-benchmark/dashboard",
	}

	// Get executable directory
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "../dashboard"),
			filepath.Join(exeDir, "../../dpb-benchmark/dashboard"),
		)
	}

	// Get working directory
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "dashboard"),
			filepath.Join(wd, "dpb-benchmark/dashboard"),
		)
	}

	for _, dir := range candidates {
		absDir, _ := filepath.Abs(dir)
		indexPath := filepath.Join(absDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return absDir
		}
	}

	return ""
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("  %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Start()
}
