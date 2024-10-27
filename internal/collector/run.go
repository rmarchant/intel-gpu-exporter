package collector

import (
	"context"
	"fmt"
	igt "github.com/clambin/intel-gpu-exporter/pkg/intel-gpu-top"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	version = "change-me"
)

func Run(ctx context.Context, r prometheus.Registerer, scanInterval time.Duration, logger *slog.Logger) error {
	logger.Info("intel-gpu-exporter starting", "version", version)
	defer logger.Info("intel-gpu-exporter shutting down")

	cmd, output, err := startTop(ctx, scanInterval, logger)
	if err != nil {
		return fmt.Errorf("failed to start intel_gpu_top: %w", err)
	}
	defer func() {
		err := cmd.Wait()
		logger.Debug("intel-gpu-top exited", "err", err)
	}()

	logger.Debug("intel-gpu-exporter is running")

	return run(ctx, r, output, logger)
}

func run(ctx context.Context, r prometheus.Registerer, output io.Reader, logger *slog.Logger) error {
	var c Collector
	c.Aggregator.Logger = logger
	r.MustRegister(&c)

	errCh := make(chan error)
	go func() {
		errCh <- c.Read(igt.V118toV117{Reader: output})
	}()

	logger.Debug("collector is running")
	defer logger.Debug("collector is shutting down")

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

func startTop(ctx context.Context, scanInterval time.Duration, logger *slog.Logger) (*exec.Cmd, io.ReadCloser, error) {
	cmdline := buildCommand(scanInterval)
	logger.Debug("top command built", "duration", scanInterval, "cmd", strings.Join(cmdline, " "))
	cmd := exec.CommandContext(ctx, cmdline[0], cmdline[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout pipe failed: %w", err)
	}
	return cmd, stdout, cmd.Start()
}

func buildCommand(scanInterval time.Duration) []string {
	const gpuTopCommand = "intel_gpu_top -J -s"

	return append(
		strings.Split(gpuTopCommand, " "),
		strconv.Itoa(int(scanInterval.Milliseconds())),
	)

}
