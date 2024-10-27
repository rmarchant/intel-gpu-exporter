package collector

import (
	"context"
	"github.com/clambin/intel-gpu-exporter/pkg/intel-gpu-top/testutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func Test_Main(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	r := prometheus.NewRegistry()
	top := testutil.FakeServer(ctx, []byte(testutil.SinglePayload), 5, false, false, 0)
	go Run(ctx, r, top)

	assert.Eventually(t, func() bool {
		_, err := http.Get("http://localhost:9090/metrics")
		return err == nil
	}, time.Second*5, time.Millisecond*200)

	cancel()
}
