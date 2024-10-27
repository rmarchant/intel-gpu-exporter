package collector

import (
	"context"
	"errors"
	"flag"
	igt "github.com/clambin/intel-gpu-exporter/pkg/intel-gpu-top"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log/slog"
	"net/http"
	"os"
)

var (
	version = "change-me"
	debug   = flag.Bool("debug", false, "Enable debug logging")
	addr    = flag.String("addr", ":9090", "Prometheus metrics listener address")
)

func Run(ctx context.Context, r prometheus.Registerer, top io.Reader) {
	flag.Parse()

	var handlerOpts slog.HandlerOptions
	if *debug {
		handlerOpts.Level = slog.LevelDebug
	}
	l := slog.New(slog.NewTextHandler(os.Stderr, &handlerOpts))

	l.Info("intel-gpu-exporter starting", "version", version, "addr", *addr)

	var c Collector
	r.MustRegister(&c)

	go func() {
		if err := c.Read(igt.V118toV117{Reader: top}); err != nil {
			l.Error("intel_gpu_top read failed", "err", err)
			os.Exit(1)
		}
	}()

	l.Debug("reader started")

	http.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	go func() {
		if err := http.ListenAndServe(*addr, nil); !errors.Is(err, http.ErrServerClosed) {
			l.Error("failed to start metrics server", "err", err)
			os.Exit(1)
		}
	}()

	l.Debug("metrics server started")
	<-ctx.Done()
	l.Info("intel-gpu-exporter shutting down")
}
