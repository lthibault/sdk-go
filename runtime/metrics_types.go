package runtime

import (
	"sync"
	"time"
)

type MetricType int

const (
	MetricPoint MetricType = iota
	MetricCounter
	MetricEWMA
	MetricGauge
	MetricHistogram
	MetricMeter
	MetricTimer
)

func (mt MetricType) String() string {
	return [...]string{"point", "counter", "ewma", "gauge", "histogram", "meter", "timer"}[mt]
}

var pools = func() (p [7]sync.Pool) {
	for i := range p {
		p[i].New = func() interface{} {
			return &Metric{Type: MetricType(i), Measures: make(map[string]interface{}, 1)}
		}
	}
	return p
}()

type Metric struct {
	Timestamp int64                  `json:"ts"`
	Type      MetricType             `json:"t"`
	Name      string                 `json:"n"`
	Measures  map[string]interface{} `json:"m"`
}

func (m *Metric) Release() {
	pools[m.Type].Put(m)
}

func NewMetric(name string, i interface{}) *Metric {
	var (
		m  *Metric
		t  MetricType
		ts = time.Now().UnixNano()
	)

	switch v := i.(type) {
	case Point:
		t = MetricPoint
		m = pools[t].Get().(*Metric)
		m.Measures["value"] = v

	case Counter:
		t = MetricCounter
		m = pools[t].Get().(*Metric)
		s := v.Snapshot()
		m.Measures["count"] = s.Count()

	case EWMA:
		t = MetricEWMA
		m = pools[t].Get().(*Metric)
		s := v.Snapshot()
		m.Measures["rate"] = s.Rate()

	case Gauge:
		t = MetricGauge
		m = pools[t].Get().(*Metric)
		s := v.Snapshot()
		m.Measures["value"] = s.Value()

	case Histogram:
		t = MetricHistogram
		m = pools[t].Get().(*Metric)
		s := v.Snapshot()
		p := s.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
		m.Measures["count"] = float64(s.Count())
		m.Measures["max"] = float64(s.Max())
		m.Measures["mean"] = s.Mean()
		m.Measures["min"] = float64(s.Min())
		m.Measures["stddev"] = s.StdDev()
		m.Measures["variance"] = s.Variance()
		m.Measures["p50"] = p[0]
		m.Measures["p75"] = p[1]
		m.Measures["p95"] = p[2]
		m.Measures["p99"] = p[3]
		m.Measures["p999"] = p[4]
		m.Measures["p9999"] = p[5]

	case Meter:
		t = MetricMeter
		m = pools[t].Get().(*Metric)
		s := v.Snapshot()
		m.Measures["count"] = float64(s.Count())
		m.Measures["m1"] = s.Rate1()
		m.Measures["m5"] = s.Rate5()
		m.Measures["m15"] = s.Rate15()
		m.Measures["mean"] = s.RateMean()

	case Timer:
		t = MetricTimer
		m = pools[t].Get().(*Metric)
		s := v.Snapshot()
		p := s.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
		m.Measures["count"] = float64(s.Count())
		m.Measures["max"] = float64(s.Max())
		m.Measures["mean"] = s.Mean()
		m.Measures["min"] = float64(s.Min())
		m.Measures["stddev"] = s.StdDev()
		m.Measures["variance"] = s.Variance()
		m.Measures["p50"] = p[0]
		m.Measures["p75"] = p[1]
		m.Measures["p95"] = p[2]
		m.Measures["p99"] = p[3]
		m.Measures["p999"] = p[4]
		m.Measures["p9999"] = p[5]
		m.Measures["m1"] = s.Rate1()
		m.Measures["m5"] = s.Rate5()
		m.Measures["m15"] = s.Rate15()
		m.Measures["meanrate"] = s.RateMean()

	default:
		panic("unexpected metric type")

	}

	m.Timestamp = ts
	m.Type = t
	m.Name = name
	return m
}