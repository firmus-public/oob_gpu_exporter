package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/firmus-public/oob_gpu_exporter/internal/collector"
	"github.com/firmus-public/oob_gpu_exporter/internal/log"
	"github.com/firmus-public/oob_gpu_exporter/internal/version"
)

const (
	contentTypeHeader     = "Content-Type"
	contentEncodingHeader = "Content-Encoding"
	acceptEncodingHeader  = "Accept-Encoding"
)

var gzipPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

const landingPageTemplate = `<html lang="en">
<head><title>Out-of-band GPU Exporter</title></head>
<body style="font-family: sans-serif">
<h2>Out-of-band GPU Exporter</h2>
<div>Build information: version=%s revision=%s</div>
<ul><li><a href="/metrics">Metrics</a> (needs <code>target</code> parameter)</li></ul>
</body>
</html>
`

func rootHandler(rsp http.ResponseWriter, req *http.Request) {
	_, err := fmt.Fprintf(rsp, landingPageTemplate, version.Version, version.Revision)
	if err != nil {
		log.Error("Error writing landing page to client %s: %v", req.Host, err)
		http.Error(rsp, "Error writing landing page to client", http.StatusInternalServerError)
		return
	}
}

func healthHandler(rsp http.ResponseWriter, req *http.Request) {
	// just return a simple 200 for now
}

func resetHandler(rsp http.ResponseWriter, req *http.Request) {
	target := req.URL.Query().Get("target")
	if target == "" {
		log.Error("Received request from %s without 'target' parameter", req.Host)
		http.Error(rsp, "Query parameter 'target' is mandatory", http.StatusBadRequest)
		return
	}

	log.Debug("Handling reset-request from %s for host %s", req.Host, target)

	collector.Reset(target)
}

func metricsHandler(rsp http.ResponseWriter, req *http.Request) {
	// Config is reloaded in the background watcher, just use current config
	target := req.URL.Query().Get("target")
	if target == "" {
		log.Error("Received request from %s without 'target' parameter", req.Host)
		http.Error(rsp, "Query parameter 'target' is mandatory", http.StatusBadRequest)
		return
	}

	log.Debug("Handling request from %s for host %s", req.Host, target)

	c, err := collector.GetCollector(target)
	if err != nil {
		errorMsg := fmt.Sprintf("Error instantiating metrics collector for host %s: %v", target, err)
		log.Error("%v", errorMsg)
		http.Error(rsp, errorMsg, http.StatusInternalServerError)
		return
	}

	log.Debug("Collecting metrics for host %s", target)

	metrics, err := c.Gather()
	if err != nil {
		errorMsg := fmt.Sprintf("Error collecting metrics for host %s: %v", target, err)
		log.Error("%v", errorMsg)
		http.Error(rsp, errorMsg, http.StatusInternalServerError)
		return
	}

	log.Debug("Metrics for host %s collected", target)

	header := rsp.Header()
	header.Set(contentTypeHeader, "text/plain")

	// Code inspired by the official Prometheus metrics http handler
	w := io.Writer(rsp)
	if gzipAccepted(req.Header) {
		header.Set(contentEncodingHeader, "gzip")
		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)

		gz.Reset(w)
		defer func() {
			err = gz.Close()
			if err != nil {
				log.Error("Error closing gzip writer for client %s: %v", req.Host, err)
			}
		}()

		w = gz
	}

	_, err = fmt.Fprint(w, metrics)
	if err != nil {
		log.Error("Error writing metrics to client %s: %v", req.Host, err)
		http.Error(rsp, "Error writing metrics to client", http.StatusInternalServerError)
		return
	}
}

// gzipAccepted returns whether the client will accept gzip-encoded content.
func gzipAccepted(header http.Header) bool {
	a := header.Get(acceptEncodingHeader)
	parts := strings.Split(a, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "gzip" || strings.HasPrefix(part, "gzip;") {
			return true
		}
	}
	return false
}
