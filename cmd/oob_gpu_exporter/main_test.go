package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"syscall"

	// "os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
    // Start a mock Redfish server serving content sourced from testdata/content
	contentDir := filepath.Join("testdata", "content")
	handler := fileHandler(contentDir)
	server := httptest.NewTLSServer(http.HandlerFunc(handler))
	defer server.Close()

    // Extract host and port from the server URL
	test_host, test_port, err := net.SplitHostPort(server.URL[len("https://"):])
	if err != nil {
		t.Fatalf("Failed to split host and port from URL: %v", err)
	}

    // Start the exporter
    if cmd, err := startExporter();err != nil {
        t.Fatalf("Failed to start exporter: %v", err)
    } else {
        defer stopExporter(cmd)
    }

    // Get metrics from the exporter
    resp, err := get("http://localhost:9347/metrics?target=" + net.JoinHostPort(test_host, test_port))
    if err != nil {
        t.Fatalf("Failed to get metrics: %v", err)
    }

    // Read expected metrics from file
    expectedContent, err := readTestFile("testdata", "expected.txt")
    if err != nil {
        t.Fatalf("Failed to read expected file: %v", err)
    }

    // Compare the metrics excluding go build version
    re := regexp.MustCompile(`"go[0-9]+.[0-9]+.[0-9]+`)
    if re.ReplaceAllString(resp, "") != re.ReplaceAllString(expectedContent,"") {
        t.Fatalf("Metrics do not match expected content.\nGot:\n%s\nExpected:\n%s", resp, expectedContent)
    }
}

func fileHandler(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName := filepath.Clean(r.URL.Path)
		filePath := filepath.Join(baseDir, fileName, "index.json")

		data, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
        if err != nil {
            log.Printf("Error writing response for %s: %v", r.URL.Path, err)
        }
	}
}

func readTestFile(path ...string) (string, error) {
	content := filepath.Join(path...)

	expectedBytes, err := os.ReadFile(content)
	if err != nil {
		return "", err
	}
	expectedContent := string(expectedBytes)

    return expectedContent, nil

}

func get(url string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = resp.Body.Close()
    if err != nil {
        fmt.Printf("Error closing response body for URL %s: %v", url, err)
    }

    return string(bodyBytes), nil
}

func startExporter() (*exec.Cmd, error) {
    cmd := exec.Command("go", "run", ".", "-config", "testdata/config.yml")
    
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,
    }

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        fmt.Printf("Failed to start command: %v\n", err)
        return nil, err
    }

    waitFor("http://localhost:9347/health", 10 )

    return cmd, nil
}

func waitFor(endpoint string, secs int) bool {
	for i := 0; i < secs*2; i++ {
		resp, err := http.Get(endpoint)
		if err == nil && resp.StatusCode == http.StatusOK {
			fmt.Println("HTTP server is up and running!")
			err = resp.Body.Close()
            if err != nil {
                fmt.Printf("Error closing response body for health check: %v", err)
            }
			return true
		}
		fmt.Printf("Waiting for server to start... (attempt %d)", i+1)
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func stopExporter(cmd *exec.Cmd) {
    err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
    if err != nil {
        fmt.Printf("Failed to kill process: %v", err)
        return  // Do not try to wait on a process we failed to kill
    }
    //nolint:errcheck // we just killed it, we don't care about the error
    cmd.Wait()
}
