package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"slices"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Func func(next http.Handler) http.Handler

type Middlewares []Func

func (mws Middlewares) Wrap(toWrap http.Handler) http.Handler {
	for _, mw := range slices.Backward(mws) {
		toWrap = mw(toWrap)
	}

	return toWrap
}

// LogRequest logs incoming requests, body included.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqBs, err := httputil.DumpRequest(req, true)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("failed to dump request: %v", err)))
			return
		}

		slog.Debug("received request", slog.String("request", string(reqBs)))

		next.ServeHTTP(rw, req)
	})
}

type recorder struct {
	http.ResponseWriter
	code int
}

func (r *recorder) WriteHeader(statusCode int) {
	r.code = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// LogResponse logs the response status code and time taken to serve the request.
func LogResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rr := recorder{
			ResponseWriter: rw,
		}

		start := time.Now()
		next.ServeHTTP(&rr, req)
		duration := time.Since(start)

		slog.Info("served request", slog.Int("status", rr.code), slog.Float64("duration", duration.Seconds()))
	})
}

// Recover recovers from panics that happen down the stack.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(rw, req)
		if reason := recover(); reason != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(req.Context(), "handler panicked", slog.Any("reason", reason))
		}
	})
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		inFlightReqs.WithLabelValues(req.Method, req.Pattern).Inc()
		defer inFlightReqs.WithLabelValues(req.Method, req.Pattern).Dec()

		rr := recorder{
			ResponseWriter: rw,
		}

		//start := time.Now()
		next.ServeHTTP(&rr, req)
		//duration := time.Since(start)

		responseCodes.WithLabelValues(req.Method, req.Pattern, strconv.Itoa(rr.code)).Inc()
	})
}

var (
	inFlightReqs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_reqs_in_flight",
		Help: "Current number of in flight requests.",
	}, []string{"method", "path"})

	responseCodes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_status_code_count",
		Help: "Number of response status codes per endpoint.",
	}, []string{"method", "path", "status_code"})
)

func InitMetrics() {
	if err := prometheus.DefaultRegisterer.Register(inFlightReqs); err != nil {
		slog.Error("error initializing in-flight reqs gauge", slog.Any("error", err))
	}
	if err := prometheus.DefaultRegisterer.Register(responseCodes); err != nil {
		slog.Error("error initializing response codes counter", slog.Any("error", err))
	}
}
