package intel_gpu_top

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

const statEntry = `
{
	"period":        { "duration": 21.473224, "unit": "ms" },
	"frequency":     { "requested": 0.000000, "actual": 0.000000, "unit": "MHz" },
	"interrupts":    { "count": 1257.379889, "unit": "irq/s" },
	"rc6":           { "value": 100.000000, "unit": "%" },
	"power":         { "GPU": 0.000000, "Package": 35.276832, "unit": "W" },
	"imc-bandwidth": { "reads": 1332.134597, "writes": 129.066989, "unit": "MiB/s" },
	"engines":    {
		"Render/3D":    { "busy": 1.000000, "sema": 0.000000, "wait": 0.000000, "unit": "%" },
		"Blitter":      { "busy": 2.000000, "sema": 0.000000, "wait": 0.000000, "unit": "%" }, 
        "Video":        { "busy": 3.000000, "sema": 0.000000, "wait": 0.000000, "unit": "%" },
		"VideoEnhance": { "busy": 4.000000, "sema": 0.000000, "wait": 0.000000, "unit": "%" }
	},
	"clients": {
        "1": {}
    }
}`

const realPayload = `
{
	"period": {
		"duration": 19.265988,
		"unit": "ms"
	},
	"frequency": {
		"requested": 0.000000,
		"actual": 0.000000,
		"unit": "MHz"
	},
	"interrupts": {
		"count": 0.000000,
		"unit": "irq/s"
	},
	"rc6": {
		"value": 0.000000,
		"unit": "%"
	},
	"power": {
		"GPU": 0.000000,
		"Package": 16.597290,
		"unit": "W"
	},
	"imc-bandwidth": {
		"reads": 1458.945797,
		"writes": 198.603567,
		"unit": "MiB/s"
	},
	"engines": {
		"Render/3D": {
			"busy": 0.000000,
			"sema": 0.000000,
			"wait": 0.000000,
			"unit": "%"
		},
		"Blitter": {
			"busy": 0.000000,
			"sema": 0.000000,
			"wait": 0.000000,
			"unit": "%"
		},
		"Video": {
			"busy": 3.000000,
			"sema": 0.000000,
			"wait": 0.000000,
			"unit": "%"
		},
		"VideoEnhance": {
			"busy": 0.000000,
			"sema": 0.000000,
			"wait": 0.000000,
			"unit": "%"
		}
	},
	"clients": {
		"4293539623": {
			"name": "Plex Transcoder",
			"pid": "1427673",
			"engine-classes": {
				"Render/3D": {
					"busy": "0.000000",
					"unit": "%"
				},
				"Blitter": {
					"busy": "0.000000",
					"unit": "%"
				},
				"Video": {
					"busy": "0.000000",
					"unit": "%"
				},
				"VideoEnhance": {
					"busy": "0.000000",
					"unit": "%"
				}
			}
		}
	}
}
`

func TestReadGPUStats(t *testing.T) {
	const count = 2
	r := &JSONFixer{Reader: statWriter(count, []byte(realPayload), time.Millisecond)}
	var found int
	for gpuStat, err := range ReadGPUStats(r) {
		require.NoError(t, err)
		assert.Equal(t, 3.0, gpuStat.Engines["Video"].Busy)
		found++
	}
	assert.Equal(t, count, found)
}

func TestAggregator(t *testing.T) {
	var a Aggregator
	assert.NoError(t, a.Read(&JSONFixer{Reader: statWriter(2, []byte(statEntry), time.Millisecond)}))

	// a.Read works asynchronously. Wait for all data to be read.
	assert.Eventually(t, func() bool { return len(a.EngineStats()) == 4 }, time.Second, time.Millisecond)

	engineStats := a.EngineStats()
	require.Len(t, engineStats, 4)
	for i, engineName := range []string{"Render/3D", "Blitter", "Video", "VideoEnhance"} {
		assert.Contains(t, engineStats, engineName)
		assert.Equal(t, float64(i+1), engineStats[engineName].Busy)
		assert.Equal(t, "%", engineStats[engineName].Unit)
	}

	assert.Equal(t, 1.0, a.ClientStats())
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

func statWriter(count int, payload []byte, interval time.Duration) io.Reader {
	r, w := io.Pipe()
	go func() {
		_, _ = w.Write([]byte("[ \n\n"))
		for range count {
			//if i != 0 {
			//	_, _ = w.Write([]byte(",\n"))
			//}
			_, _ = w.Write(payload)
			time.Sleep(interval)
		}
		_, _ = w.Write([]byte("\n]\n"))
		_ = w.Close()
	}()
	return r
}

func BenchmarkReadGPUStats(b *testing.B) {
	for range b.N {
		r := &JSONFixer{Reader: statWriter(10, []byte(statEntry), time.Millisecond)}
		var a Aggregator
		if err := a.Read(r); err != nil {
			b.Fatal(err)
		}
	}
}
