package main

import (
	"context"
	"fmt"
	"github.com/clambin/intel-gpu-exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	topCmd, topOutput, err := startTop(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "intel_gpu_top failed to start: %s", err.Error())
		os.Exit(1)
	}

	go collector.Run(ctx, prometheus.DefaultRegisterer, topOutput)
	_ = topCmd.Wait()
}

const gpuTopCommand = "intel_gpu_top -J -s 5000"

func startTop(ctx context.Context) (*exec.Cmd, io.ReadCloser, error) {
	cmdline := strings.Split(gpuTopCommand, " ")
	cmd := exec.CommandContext(ctx, cmdline[0], cmdline[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout pipe failed: %w", err)
	}
	return cmd, stdout, cmd.Start()
}
