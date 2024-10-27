package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	engineMetric = prometheus.NewDesc(
		prometheus.BuildFQName("gpumon", "engine", "usage"),
		"Usage statistics for the different GPU engines",
		[]string{"engine", "attrib"},
		nil,
	)
	powerMetric = prometheus.NewDesc(
		prometheus.BuildFQName("gpumon", "", "power"),
		"Power consumption by type",
		[]string{"type"},
		nil,
	)
	clientMetric = prometheus.NewDesc(
		prometheus.BuildFQName("gpumon", "clients", "count"),
		"Number of active clients",
		nil,
		nil,
	)
)

type Collector struct {
	Aggregator
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- engineMetric
	ch <- powerMetric
	ch <- clientMetric
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	for engine, engineStats := range c.EngineStats() {
		ch <- prometheus.MustNewConstMetric(engineMetric, prometheus.GaugeValue, engineStats.Busy, engine, "busy")
		ch <- prometheus.MustNewConstMetric(engineMetric, prometheus.GaugeValue, engineStats.Sema, engine, "sema")
		ch <- prometheus.MustNewConstMetric(engineMetric, prometheus.GaugeValue, engineStats.Wait, engine, "wait")
	}
	gpuPower, packagePower := c.PowerStats()
	ch <- prometheus.MustNewConstMetric(powerMetric, prometheus.GaugeValue, packagePower, "pkg")
	ch <- prometheus.MustNewConstMetric(powerMetric, prometheus.GaugeValue, gpuPower, "gpu")
	ch <- prometheus.MustNewConstMetric(clientMetric, prometheus.GaugeValue, c.ClientStats())
	c.Reset()
}
