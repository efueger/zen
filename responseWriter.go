package zen

import "net/http"

// -----------------------------------------------------------------------------
// Simple wrapper around a ResponseWriter

// responseWriter is a wrapper for the http.ResponseWriter
// to track if response was written to. It also allows us
// to automatically set certain headers, such as Content-Type,
// Access-Control-Allow-Origin, etc.
type responseWriter struct {
	writer  http.ResponseWriter
	written bool
}

// Header returns the header map that will be sent by WriteHeader.
func (w *responseWriter) Header() http.Header {
	return w.writer.Header()
}

// Write writes the data to the connection as part of an HTTP reply,
// and sets `written` to true
func (w *responseWriter) Write(p []byte) (int, error) {
	w.written = true
	return w.writer.Write(p)
}

// WriteHeader sends an HTTP response header with status code,
// and sets `written` to true
func (w *responseWriter) WriteHeader(code int) {
	w.written = true
	w.writer.WriteHeader(code)
}
