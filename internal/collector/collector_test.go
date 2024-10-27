package collector

import (
	"context"
	testutil2 "github.com/clambin/intel-gpu-exporter/pkg/intel-gpu-top/testutil"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var c Collector
	go func() {
		require.NoError(t, c.Read(testutil2.FakeServer(ctx, []byte(testutil2.SinglePayload), 1, false, false, 0)))
	}()

	// wait for the aggregator to read in the data
	assert.Eventually(t, func() bool { return c.len() > 0 }, time.Second, time.Millisecond)

	assert.NoError(t, testutil.CollectAndCompare(&c, strings.NewReader(`
# HELP gpumon_clients_count Number of active clients
# TYPE gpumon_clients_count gauge
gpumon_clients_count 1

# HELP gpumon_engine_usage Usage statistics for the different GPU engines
# TYPE gpumon_engine_usage gauge
gpumon_engine_usage{attrib="busy",engine="Blitter"} 2
gpumon_engine_usage{attrib="busy",engine="Render/3D"} 1
gpumon_engine_usage{attrib="busy",engine="Video"} 3
gpumon_engine_usage{attrib="busy",engine="VideoEnhance"} 4
gpumon_engine_usage{attrib="sema",engine="Blitter"} 0
gpumon_engine_usage{attrib="sema",engine="Render/3D"} 0
gpumon_engine_usage{attrib="sema",engine="Video"} 0
gpumon_engine_usage{attrib="sema",engine="VideoEnhance"} 0
gpumon_engine_usage{attrib="wait",engine="Blitter"} 0
gpumon_engine_usage{attrib="wait",engine="Render/3D"} 0
gpumon_engine_usage{attrib="wait",engine="Video"} 0
gpumon_engine_usage{attrib="wait",engine="VideoEnhance"} 0

# HELP gpumon_power Power consumption by type
# TYPE gpumon_power gauge
gpumon_power{type="gpu"} 1
gpumon_power{type="pkg"} 4
`)))
}
