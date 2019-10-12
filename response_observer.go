package main

import "net/http"

type responseWriterObserver struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	size        int64
}

func (s *responseWriterObserver) Write(data []byte) (int, error) {
	if !s.wroteHeader {
		s.WriteHeader(http.StatusOK)
	}
	n, err := s.ResponseWriter.Write(data)
	s.size += int64(n)
	return n, err
}
func (s *responseWriterObserver) WriteHeader(statusCode int) {
	s.wroteHeader = true
	s.statusCode = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}
