package collector

import (
	"context"
	intel_gpu_top "github.com/clambin/intel-gpu-exporter/pkg/intel-gpu-top"
	"github.com/clambin/intel-gpu-exporter/pkg/intel-gpu-top/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestAggregator(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	const payloadCount = 4
	r := testutil.FakeServer(ctx, []byte(testutil.SinglePayload), payloadCount, false, false, time.Millisecond)
	var a Aggregator
	a.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	assert.NoError(t, a.Read(r))

	// a.Read works asynchronously. Wait for all data to be read.
	assert.Eventually(t, func() bool { return len(a.EngineStats()) == payloadCount }, time.Second, time.Millisecond)

	wantEngines := []string{"Render/3D", "Blitter", "Video", "VideoEnhance"}

	engineStats := a.EngineStats()
	require.Len(t, engineStats, len(wantEngines))
	for i, engineName := range wantEngines {
		assert.Contains(t, engineStats, engineName)
		assert.Equal(t, float64(i+1), engineStats[engineName].Busy)
		assert.Equal(t, "%", engineStats[engineName].Unit)
	}

	assert.Equal(t, 1.0, a.ClientStats())
	gpu, pkg := a.PowerStats()
	assert.Equal(t, 1.0, gpu)
	assert.Equal(t, 4.0, pkg)

	cancel()
}

func TestAggregator_Reset(t *testing.T) {
	var a Aggregator
	a.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	assert.Len(t, a.stats, 0)
	a.Reset()
	assert.Len(t, a.stats, 0)
	var stat intel_gpu_top.GPUStats
	for i := range 5 {
		stat.Power.GPU = float64(i)
		a.add(stat)
	}
	assert.Len(t, a.stats, 5)
	a.Reset()
	require.Len(t, a.stats, 1)
	assert.Equal(t, 4.0, a.stats[0].Power.GPU)

}

func TestEngineStats_LogValue(t *testing.T) {
	stats := EngineStats{
		"FOO": {},
		"BAR": {},
	}
	assert.Equal(t, "BAR,FOO", stats.LogValue().String())
	clear(stats)
	assert.Equal(t, "", stats.LogValue().String())
}

func Test_medianFunc(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		values := make([]float64, 5)
		for i := range 5 {
			values[i] = float64(i)
		}
		assert.Equal(t, float64(2), medianFunc(values, func(f float64) float64 { return f }))
		values = make([]float64, 6)
		for i := range 6 {
			values[i] = float64(i)
		}
		assert.Equal(t, 2.5, medianFunc(values, func(f float64) float64 { return f }))

		assert.Zero(t, medianFunc(nil, func(f float64) float64 { return f }))
	})
}

// Before:
// Benchmark_mediaFunc-16            331898              3490 ns/op            8192 B/op          1 allocs/op
func Benchmark_mediaFunc(b *testing.B) {
	const count = 1001
	values := make([]float64, count)
	for i := range values {
		values[i] = float64(i)
	}
	b.ResetTimer()
	for range b.N {
		if value := medianFunc(values, func(f float64) float64 { return f }); value != count/2 {
			b.Fatalf("expected %f, got %f", float64(count/2), value)
		}
	}
}

// Current:
// BenchmarkAggregator_EngineStats-16          4962            234863 ns/op          457877 B/op         58 allocs/op
// After:
// BenchmarkAggregator_EngineStats-16          4868            234483 ns/op          385557 B/op         26 allocs/op
func BenchmarkAggregator_EngineStats(b *testing.B) {
	const count = 1001
	var engineNames = []string{"Render/3D", "Blitter", "Video", "VideoEnhance"}
	var a Aggregator
	a.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	for range count {
		var stats intel_gpu_top.GPUStats
		stats.Engines = make(map[string]intel_gpu_top.EngineStats, len(engineNames))
		for _, engine := range engineNames {
			stats.Engines[engine] = intel_gpu_top.EngineStats{}
		}
		a.add(stats)
	}
	b.ResetTimer()
	for range b.N {
		if stats := a.EngineStats(); len(stats) != len(engineNames) {
			b.Fatalf("expected %d engines, got %d", len(engineNames), len(stats))
		}
	}
}
