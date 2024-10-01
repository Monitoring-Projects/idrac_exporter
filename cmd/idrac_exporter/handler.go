package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/mrlhansen/idrac_exporter/internal/collector"
	"github.com/mrlhansen/idrac_exporter/internal/log"
)

const (
	contentTypeHeader     = "Content-Type"
	contentEncodingHeader = "Content-Encoding"
	acceptEncodingHeader  = "Accept-Encoding"
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

func HealthHandler(rsp http.ResponseWriter, req *http.Request) {
	// just return a simple 200 for now
}

func ResetHandler(rsp http.ResponseWriter, req *http.Request) {
	target := req.URL.Query().Get("target")
	if target == "" {
		log.Error("Received request from %s without 'target' parameter", req.Host)
		http.Error(rsp, "Query parameter 'target' is mandatory", http.StatusBadRequest)
		return
	}

	log.Debug("Handling reset-request from %s for host %s", req.Host, target)

	collector.Reset(target)
}

func MetricsHandler(rsp http.ResponseWriter, req *http.Request) {
	target := req.URL.Query().Get("target")
	if target == "" {
		log.Error("Received request from %s without 'target' parameter", req.Host)
		http.Error(rsp, "Query parameter 'target' is mandatory", http.StatusBadRequest)
		return
	}

	log.Debug("Handling request from %s for host %s", req.Host, target)

	c, err := collector.GetCollector(target)
	if err != nil {
		errorMsg := fmt.Sprintf("Error instantiating metrics collector for host %s: %v\n", target, err)
		log.Error(errorMsg)
		http.Error(rsp, errorMsg, http.StatusInternalServerError)
		return
	}

	log.Debug("Collecting metrics for host %s", target)

	metrics, err := c.Gather()
	if err != nil {
		errorMsg := fmt.Sprintf("Error collecting metrics for host %s: %v\n", target, err)
		log.Error(errorMsg)
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
		defer gz.Close()

		w = gz
	}

	fmt.Fprint(w, metrics)
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
