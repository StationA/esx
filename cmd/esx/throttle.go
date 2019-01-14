package main

import (
	"math"
	"sync"
	"time"
)

type SamplingThrottle struct {
	mut           sync.Mutex
	limit         time.Duration
	windowSize    int
	backoffFactor float64
	samples       []float64
}

type ThrottleOpt func(s *SamplingThrottle)

func NewSamplingThrottle(opts ...ThrottleOpt) *SamplingThrottle {
	s := new(SamplingThrottle)
	// Set defaults
	s.limit = 30 * time.Second
	s.windowSize = 10
	s.backoffFactor = 1.0
	// Then optionally override
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func SetLimit(limit time.Duration) ThrottleOpt {
	return func(s *SamplingThrottle) {
		s.limit = limit
	}
}

func SetWindowSize(size int) ThrottleOpt {
	return func(s *SamplingThrottle) {
		s.windowSize = size
	}
}

func SetBackoffFactor(factor float64) ThrottleOpt {
	return func(s *SamplingThrottle) {
		s.backoffFactor = factor
	}
}

func (s *SamplingThrottle) Collect(t time.Duration) {
	defer s.mut.Unlock()
	s.mut.Lock()
	// Shift sliding window forward if it's already full
	if len(s.samples) == s.windowSize {
		s.samples = s.samples[1:]
	}
	s.samples = append(s.samples, float64(t.Seconds()))
}

func (s *SamplingThrottle) computeOLS() (bool, float64, float64) {
	if len(s.samples) >= s.windowSize {
		num := 0.0
		den := 0.0
		xs := make([]float64, len(s.samples))
		for i := 0; i < len(s.samples); i++ {
			xs[i] = float64(i)
		}
		xavg := computeAvg(xs)
		yavg := computeAvg(s.samples)
		for x, y := range s.samples {
			xdelta := float64(x) - xavg
			num += xdelta * (y - yavg)
			den += math.Pow(xdelta, 2)
		}
		m := num / den
		b := yavg - m*xavg
		return true, m, b
	}
	return false, 0, 0
}

func (s *SamplingThrottle) computeBackoff(m, b float64) (time.Duration, time.Duration) {
	var backoffTime time.Duration
	// Predict the duration of the next call
	x := float64(len(s.samples))
	pred := m*x + b
	strength := math.Pow(pred, 2.0) / s.limit.Seconds()
	backoff := strength * s.backoffFactor
	if backoff > 0 {
		backoffTime = time.Duration(backoff * float64(time.Second))
	}
	return time.Duration(pred * float64(time.Second)), backoffTime
}

func (s *SamplingThrottle) Wait() <-chan time.Time {
	defer s.mut.Unlock()
	s.mut.Lock()
	log := Log.WithField("proc", "throttle")
	var backoff time.Duration
	ready, m, b := s.computeOLS()
	if ready {
		log.Debugf("y = %.4fx + %.4f", m, b)
		pred, backoff := s.computeBackoff(m, b)
		log.Debugf("pred = %v, backoff = %v", pred, backoff)
		if backoff > 0 {
			log.Warnf("Throttling worker for %.2fs", backoff.Seconds())
		}
	}
	return time.After(backoff)
}

func computeAvg(samples []float64) float64 {
	size := float64(len(samples))
	sum := 0.0
	for _, s := range samples {
		sum += s
	}
	avg := sum / size
	return avg
}
