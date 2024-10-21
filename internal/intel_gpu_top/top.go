package intel_gpu_top

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"slices"
	"sync"
)

type GPUStats struct {
	Period struct {
		Duration float64 `json:"duration"`
		Unit     string  `json:"unit"`
	} `json:"period"`
	Frequency struct {
		Requested float64 `json:"requested"`
		Actual    float64 `json:"actual"`
		Unit      string  `json:"unit"`
	} `json:"frequency"`
	Interrupts struct {
		Count float64 `json:"count"`
		Unit  string  `json:"unit"`
	} `json:"interrupts"`
	Rc6 struct {
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
	} `json:"rc6"`
	Power struct {
		GPU     float64 `json:"GPU"`
		Package float64 `json:"Package"`
		Unit    string  `json:"unit"`
	} `json:"power"`
	ImcBandwidth struct {
		Reads  float64 `json:"reads"`
		Writes float64 `json:"writes"`
		Unit   string  `json:"unit"`
	} `json:"imc-bandwidth"`
	Engines map[string]EngineStats `json:"engines"`
	Clients map[string]Client      `json:"clients"`
}

type EngineStats struct {
	Busy float64 `json:"busy"`
	Sema float64 `json:"sema"`
	Wait float64 `json:"wait"`
	Unit string  `json:"unit"`
}

type Client struct {
	Name          string `json:"name"`
	Pid           string `json:"pid"`
	EngineClasses map[string]struct {
		Busy string `json:"busy"`
		Unit string `json:"unit"`
	} `json:"engine-classes"`
}

func ReadGPUStats(r io.Reader) iter.Seq2[GPUStats, error] {
	return func(yield func(GPUStats, error) bool) {
		dec := json.NewDecoder(r)
		token, err := dec.Token()
		if err != nil {
			yield(GPUStats{}, fmt.Errorf("failed to read opening token: %w", err))
			return
		}
		if delim, ok := token.(json.Delim); !ok || delim != '[' {
			yield(GPUStats{}, fmt.Errorf("expected opening bracket but got %v", token))
			return
		}
		for dec.More() {
			var stats GPUStats
			if err = dec.Decode(&stats); err == nil {
				if !yield(stats, nil) {
					return
				}
			} else {
				break
			}
		}
		_, _ = dec.Token()
		if err != nil && !errors.Is(err, io.EOF) {
			yield(GPUStats{}, fmt.Errorf("GetGPUStats: %w", err))
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Aggregator struct {
	stats []GPUStats
	lock  sync.RWMutex
}

func (a *Aggregator) Read(r io.Reader) error {
	for stat, err := range ReadGPUStats(r) {
		if err != nil {
			return fmt.Errorf("error while reading stats: %w", err)
		}
		a.lock.Lock()
		a.stats = append(a.stats, stat)
		a.lock.Unlock()
	}
	return nil
}

func (a *Aggregator) Reset() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.stats = a.stats[:0]
}

func (a *Aggregator) PowerStats() (float64, float64) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return medianFunc(a.stats, func(stats GPUStats) float64 { return stats.Power.GPU }),
		medianFunc(a.stats, func(stats GPUStats) float64 { return stats.Power.Package })
}

func (a *Aggregator) EngineStats() map[string]EngineStats {
	a.lock.RLock()
	defer a.lock.RUnlock()

	statsByEngine := a.getEngineStatsByEngineName()

	engineStats := make(map[string]EngineStats, len(statsByEngine))

	for engine, stats := range statsByEngine {
		if len(stats) > 0 {
			engineStats[engine] = EngineStats{
				Busy: medianFunc(stats, func(stats EngineStats) float64 { return stats.Busy }),
				Sema: medianFunc(stats, func(stats EngineStats) float64 { return stats.Sema }),
				Wait: medianFunc(stats, func(stats EngineStats) float64 { return stats.Wait }),
				Unit: stats[0].Unit,
			}
		}
	}
	return engineStats
}

func (a *Aggregator) getEngineStatsByEngineName() map[string][]EngineStats {
	stats := make(map[string][]EngineStats)
	for _, stat := range a.stats {
		for engine, stat := range stat.Engines {
			stats[engine] = append(stats[engine], stat)
		}
	}
	return stats
}

func (a *Aggregator) ClientStats() float64 {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return medianFunc(a.stats, func(stats GPUStats) float64 { return float64(len(stats.Clients)) })
}

func medianFunc[T any](entries []T, f func(T) float64) float64 {
	if len(entries) == 0 {
		return 0
	}
	n := len(entries)
	values := make([]float64, 0, len(entries))
	for _, entry := range entries {
		values = append(values, f(entry))
	}
	slices.Sort(values)
	// Check if the number of elements is odd or even
	if n%2 == 1 {
		// Odd length, return the middle element
		return values[n/2]
	}
	// Even length, return the average of the two middle elements
	return (values[n/2-1] + values[n/2]) / 2
}
