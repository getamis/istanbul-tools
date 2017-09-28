// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package metrics

import (
	"fmt"

	"github.com/rcrowley/go-metrics"
)

type DefaultRegistry struct {
	registry metrics.Registry
}

func NewRegistry() *DefaultRegistry {
	r := metrics.NewRegistry()
	return &DefaultRegistry{registry: r}
}

func (r *DefaultRegistry) NewCounter(name string) metrics.Counter {
	return metrics.GetOrRegisterCounter(name, r.registry)
}

func (r *DefaultRegistry) NewMeter(name string) metrics.Meter {
	return metrics.GetOrRegisterMeter(name, r.registry)
}

func (r *DefaultRegistry) NewTimer(name string) metrics.Timer {
	return metrics.GetOrRegisterTimer(name, r.registry)
}

func (r *DefaultRegistry) NewHistogram(name string) metrics.Histogram {
	return metrics.GetOrRegisterHistogram(name, r.registry, metrics.NewExpDecaySample(1028, 0.015))
}

// -----------------------------------------------------------------
func (r *DefaultRegistry) Export() {
	r.export()
}

func (r *DefaultRegistry) export() {
	r.registry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			fmt.Printf("counter %s\n", name)
			fmt.Printf("  count:       %9d\n", metric.Count())
		case metrics.Gauge:
			fmt.Printf("gauge %s\n", name)
			fmt.Printf("  value:       %9d\n", metric.Value())
		case metrics.GaugeFloat64:
			fmt.Printf("gauge %s\n", name)
			fmt.Printf("  value:       %f\n", metric.Value())
		case metrics.Healthcheck:
			metric.Check()
			fmt.Printf("healthcheck %s\n", name)
			fmt.Printf("  error:       %v\n", metric.Error())
		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			fmt.Printf("histogram %s\n", name)
			fmt.Printf("  count:       %9d\n", h.Count())
			fmt.Printf("  min:         %9d\n", h.Min())
			fmt.Printf("  max:         %9d\n", h.Max())
			fmt.Printf("  mean:        %e\n", h.Mean())
			fmt.Printf("  stddev:      %e\n", h.StdDev())
			fmt.Printf("  median:      %e\n", ps[0])
			fmt.Printf("  75%%:         %e\n", ps[1])
			fmt.Printf("  95%%:         %e\n", ps[2])
			fmt.Printf("  99%%:         %e\n", ps[3])
			fmt.Printf("  99.9%%:       %e\n", ps[4])
		case metrics.Meter:
			m := metric.Snapshot()
			fmt.Printf("meter %s\n", name)
			fmt.Printf("  count:       %9d\n", m.Count())
			fmt.Printf("  1-min rate:  %e\n", m.Rate1())
			fmt.Printf("  5-min rate:  %e\n", m.Rate5())
			fmt.Printf("  15-min rate: %e\n", m.Rate15())
			fmt.Printf("  mean rate:   %e\n", m.RateMean())
		case metrics.Timer:
			t := metric.Snapshot()
			ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			fmt.Printf("timer %s\n", name)
			fmt.Printf("  count:       %9d\n", t.Count())
			fmt.Printf("  min:         %e\n", float64(t.Min()))
			fmt.Printf("  max:         %e\n", float64(t.Max()))
			fmt.Printf("  mean:        %e\n", t.Mean())
			fmt.Printf("  stddev:      %e\n", t.StdDev())
			fmt.Printf("  median:      %e\n", ps[0])
			fmt.Printf("  75%%:         %e\n", ps[1])
			fmt.Printf("  95%%:         %e\n", ps[2])
			fmt.Printf("  99%%:         %e\n", ps[3])
			fmt.Printf("  99.9%%:       %e\n", ps[4])
			fmt.Printf("  1-min rate:  %e\n", t.Rate1())
			fmt.Printf("  5-min rate:  %e\n", t.Rate5())
			fmt.Printf("  15-min rate: %e\n", t.Rate15())
			fmt.Printf("  mean rate:   %e\n", t.RateMean())
		}
	})
}
