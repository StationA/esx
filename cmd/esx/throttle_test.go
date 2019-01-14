package main

import (
	"math"
	"testing"
	"time"
)

const (
	// Error tolerance
	E = 0.001
)

func makeData(m float64, b float64, size int) []time.Duration {
	var data []time.Duration
	for i := 0; i < size; i++ {
		t := m*float64(i) + b
		d := time.Duration(t * float64(time.Second))
		data = append(data, d)
	}
	return data
}

func TestOLS(t *testing.T) {
	for c := 0; c < 9; c++ {
		m := float64(c + 1)
		b := float64(c + 1)
		size := 25
		data := makeData(m, b, size)
		throttle := NewSamplingThrottle(
			SetWindowSize(size),
		)
		for _, d := range data {
			throttle.Collect(d)
		}
		ready, c_m, c_b := throttle.computeOLS()
		if !ready {
			t.Fail()
		}
		m_err := math.Abs(m-c_m) / m
		b_err := math.Abs(b-c_b) / b
		y_err := 0.0
		y_pred := make([]float64, len(data))
		for i := 0; i < len(data); i++ {
			pred := c_m*float64(i) + c_b
			expected := data[i].Seconds()
			y_pred[i] = pred
			y_err += math.Abs(expected - pred)
		}
		mae := y_err / float64(len(data))

		t.Logf("f_exp(x) = %.4fx + %.4f", m, b)
		t.Logf("f_pred(x) = %.4fx + %.4f", c_m, c_b)
		t.Logf("y = %v", throttle.samples)
		t.Logf("y_pred = %v", y_pred)
		t.Logf("m_err = %.4f, b_err = %.4f, mae = %.4f", m_err, b_err, mae)
		if m_err > E || b_err > E || mae > E {
			t.Errorf("bad ols")
		}
	}
}

func TestBackoff(t *testing.T) {
	limit := 30 * time.Second
	window := 10
	numWindows := 4
	var data []time.Duration
	data = append(data, makeData(1, 0, 10)...)
	data = append(data, makeData(2, 10, 10)...)
	data = append(data, makeData(0, 30, 10)...)
	data = append(data, makeData(-3, 30, 10)...)
	throttle := NewSamplingThrottle(
		SetLimit(limit),
		SetWindowSize(window),
	)
	for w := 0; w < numWindows; w++ {
		for _, d := range data[window*w : window*(w+1)] {
			throttle.Collect(d)
		}
		t.Logf("samples[%d] = %v", w, throttle.samples)
		_, m, b := throttle.computeOLS()
		pred, backoff := throttle.computeBackoff(m, b)
		t.Logf("pred[%d] = %v, backoff[%d] = %v", w, pred, w, backoff)
	}
}
