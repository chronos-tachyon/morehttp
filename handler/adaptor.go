package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/chronos-tachyon/morehttp/response"
)

var (
	PromPanicsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "http_panics_total",
			Help: "Total number of HTTP requests that triggered a panic().",
		},
	)
	PromRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by status code.",
		},
		[]string{"code"},
	)
	PromLatencyTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_latency_total",
			Help: "Total number of wall-time seconds spent on handling HTTP responses by status code.",
		},
		[]string{"code"},
	)
	PromRecvBytesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_recv_bytes_total",
			Help: "Total number of bytes received in HTTP requests by status code.",
		},
		[]string{"code"},
	)
	PromSendBytesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_send_bytes_total",
			Help: "Total number of bytes sent in HTTP responses by status code.",
		},
		[]string{"code"},
	)
	PromLatencyHist = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_latency_seconds_hist",
			Help:    "Histogram of HTTP request latency.",
			Buckets: []float64{0.005, 0.010, 0.020, 0.050, 0.100, 0.200, 0.500, 1.000, 2.000, 5.000, 10.000},
		},
	)
	PromRecvBytesHist = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_recv_bytes_hist",
			Help:    "Histogram of HTTP request size.",
			Buckets: prometheus.ExponentialBuckets(1024, 4.0, 6),
		},
	)
	PromSendBytesHist = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_send_bytes_hist",
			Help:    "Histogram of HTTP response size.",
			Buckets: prometheus.ExponentialBuckets(1024, 4.0, 6),
		},
	)
)

type Adaptor struct {
	Inner Handler
}

func (a Adaptor) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ww := response.NewWriter(w, req)
	startTime := time.Now()

	defer func() {
		panicValue := recover()
		code := strconv.Itoa(ww.Status())
		elapsedDuration := time.Since(startTime)
		elapsed := float64(elapsedDuration) / float64(time.Second)
		recvBytes := float64(0.0) // FIXME
		sendBytes := float64(ww.BytesWritten())
		labels := prometheus.Labels{"code": code}

		ww.MaybeWriteHeader(http.StatusInternalServerError)
		if panicValue != nil {
			PromPanicsTotal.Inc()
		}
		PromRequestsTotal.With(labels).Inc()
		PromLatencyTotal.With(labels).Add(elapsed)
		PromRecvBytesTotal.With(labels).Add(recvBytes)
		PromSendBytesTotal.With(labels).Add(sendBytes)
		PromLatencyHist.Observe(elapsed)
		PromRecvBytesHist.Observe(recvBytes)
		PromSendBytesHist.Observe(sendBytes)

		if panicValue != nil {
			err, ok := panicValue.(error)
			if !ok {
				err = PanicError{Value: panicValue}
			}
			OnPanic(err)
		}
	}()

	resp := a.Inner.Handle(req)
	err := resp.Serve(ww)
	if err != nil {
		panic(err)
	}
}

var _ http.Handler = Adaptor{}
