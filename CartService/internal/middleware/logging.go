package middleware

import (
	"log"
	"net/http"
	"time"
)

type Writer struct {
	http.ResponseWriter
	status int
}

func (w *Writer) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &Writer{ResponseWriter: w, status: 200}
		start := time.Now()
		next.ServeHTTP(sw, r)
		d := time.Since(start)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, sw.status, d)
	})
}
