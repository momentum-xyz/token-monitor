package tokensvc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/OdysseyMomentumExperience/token-service/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/ory/x/errorsx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
)

type PrometheusServer struct {
	server *http.Server
}

func NewPrometheusServer() *PrometheusServer {
	r := chi.NewRouter()
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	server := &http.Server{
		Addr:    "0.0.0.0:2112",
		Handler: r,
	}

	return &PrometheusServer{
		server: server,
	}
}

func (s *PrometheusServer) Start(ctx context.Context) {
	go func() {
		log.Error(errorsx.WithStack(s.server.ListenAndServe()))
	}()

	go func() {
		<-ctx.Done()
		log.Error(s.server.Shutdown(ctx))
	}()
}

func DumpMetrics(path string, format expfmt.Format) error {
	w := new(bytes.Buffer)
	enc := expfmt.NewEncoder(w, format)
	metrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		enc.Encode(m)
	}
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, w.String())
	if err != nil {
		return err
	}

	return nil
}
