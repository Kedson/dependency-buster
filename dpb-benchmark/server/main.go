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
	"time"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	open := flag.Bool("open", true, "Open browser automatically")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *version {
		fmt.Printf("dependency-buster Dashboard Server\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// Find dashboard directory (relative to this binary or workspace)
	dashboardDir := findDashboardDir()
	if dashboardDir == "" {
		log.Fatal("Could not find dashboard directory")
	}

	// Create file server
	fs := http.FileServer(http.Dir(dashboardDir))
	http.Handle("/", fs)

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost%s", addr)

	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│  dependency-buster // Dashboard Server                  │")
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│  ▶ Version:  %-42s│\n", Version)
	fmt.Printf("│  ▶ Serving:  %-42s│\n", dashboardDir)
	fmt.Printf("│  ▶ URL:      %-42s│\n", url)
	fmt.Printf("│  ▶ Started:  %-42s│\n", time.Now().Format("2006-01-02 15:04:05"))
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
