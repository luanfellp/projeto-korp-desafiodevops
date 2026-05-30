package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Response representa a estrutura da resposta JSON do endpoint /projeto-korp
type Response struct {
	Nome    string `json:"nome"`
	Horario string `json:"horario"`
}

var (
	// httpRequestsTotal — volume de requisições (obrigatório pelo desafio)
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total de requisições HTTP recebidas pelo serviço.",
		},
		[]string{"method", "path", "status_code"},
	)

	// serviceUp — disponibilidade do serviço (obrigatório pelo desafio)
	// Vale 1 enquanto o processo está vivo; Prometheus exibe 0/ausente quando cai.
	serviceUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_up",
			Help: "Disponibilidade do serviço (1 = disponível, 0 = indisponível).",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(serviceUp)
	serviceUp.Set(1) // sinaliza que o serviço está no ar
}

// responseWriter encapsula http.ResponseWriter para capturar o status HTTP
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// metricsMiddleware registra métricas de volume por método, path e status
func metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newResponseWriter(w)
		next(rw, r)
		httpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", rw.statusCode),
		).Inc()
	}
}

// projetoKorpHandler — GET /projeto-korp
func projetoKorpHandler(w http.ResponseWriter, r *http.Request) {
	resp := Response{
		Nome:    "Projeto Korp",
		Horario: time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Erro ao serializar resposta: %v", err)
	}
}

// healthHandler — GET /health (usado pelo Prometheus para disponibilidade)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/projeto-korp", metricsMiddleware(projetoKorpHandler))
	mux.HandleFunc("/health", metricsMiddleware(healthHandler))
	mux.Handle("/metrics", promhttp.Handler()) // endpoint Prometheus

	log.Println("http-server-projeto-korp iniciado na porta :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Falha ao iniciar servidor: %v", err)
	}
}
