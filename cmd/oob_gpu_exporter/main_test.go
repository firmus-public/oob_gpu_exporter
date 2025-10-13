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

func TestDell(t *testing.T) {
	server := NewTestServer(t, "dell")
	defer server.Close()

    exporter := NewOOBGPUExporter(t, "testdata/config.yml")
	defer exporter.Stop()

    resp := getMetrics(t, server)

    assert_equal(t, "dell_expected.txt", resp)
}

func TestSupermicro(t *testing.T) {
	server := NewTestServer(t, "supermicro")
	defer server.Close()

    exporter := NewOOBGPUExporter(t, "testdata/config.yml")
	defer exporter.Stop()

    resp := getMetrics(t, server)

    assert_equal(t, "supermicro_expected.txt", resp)
}

// TestServer is a simple HTTPS server that serves files from a specified directory.

type TestServer struct {
	t   *testing.T
	server  *httptest.Server
	Host    string
	Port	string
}

func NewTestServer(t *testing.T, content string) *TestServer {
	contentDir := filepath.Join("testdata", content)
	handler := fileHandler(contentDir)
	server := httptest.NewTLSServer(http.HandlerFunc(handler))

	host, port, err := net.SplitHostPort(server.URL[len("https://"):])
	if err != nil {
		t.Fatalf("Failed to parse test server URL: %v", err)
	}

	return &TestServer{
		t: t,
		server: server,
		Host:   host,
		Port:   port,
	}
}

func (testServer *TestServer) Close() {
	testServer.server.Close()
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

// OOBGPUExporter manages the lifecycle of the oob_gpu_exporter process for testing.

type OOBGPUExporter struct {
	t   *testing.T
	cmd *exec.Cmd
}

func NewOOBGPUExporter(t *testing.T, configPath string) *OOBGPUExporter {
    cmd := exec.Command("go", "run", ".", "-config", configPath)
    
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,
    }

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        panic(fmt.Sprintf("Failed to start command: %v\n", err))
    }

    waitFor("http://localhost:9347/health", 10 )

    return &OOBGPUExporter{
		t:  t,
		cmd: cmd,
	}
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

func (oobGPUExporter *OOBGPUExporter) Stop() {
    err := syscall.Kill(-oobGPUExporter.cmd.Process.Pid, syscall.SIGINT)
    if err != nil {
        oobGPUExporter.t.Fatalf("Failed to kill process: %v", err)
        return  // Do not try to wait on a process we failed to kill
    }
    //nolint:errcheck // we just killed it, we don't care about the error
    oobGPUExporter.cmd.Wait()
}

// Convenience functions

func getMetrics(t *testing.T, server *TestServer) string {
	resp, err := get("http://localhost:9347/metrics?target=" + net.JoinHostPort(server.Host, server.Port))
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	return resp
}

func assert_equal(t *testing.T, expected string, resp string) {
	expectedContent, err := readTestFile("testdata", expected)
	if err != nil {
		t.Fatalf("Failed to read expected response: %v", err)
	}

	// Compare the metrics excluding go build version
	re := regexp.MustCompile(`"go[0-9]+.[0-9]+.[0-9]+`)
	if re.ReplaceAllString(resp, "") != re.ReplaceAllString(expectedContent, "") {
		t.Fatalf("Metrics do not match expected content.\nGot:\n%s\nExpected:\n%s", resp, expectedContent)
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

