package simpleproxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoggedMux struct {
	*http.ServeMux
	VhostLogListeners map[string]chan string
}

func NewLoggedMux() *LoggedMux {
	return &LoggedMux{
		ServeMux:          http.NewServeMux(),
		VhostLogListeners: make(map[string]chan string),
	}
}

func (mux *LoggedMux) AddLogListener(host string, listener chan string) {
	mux.VhostLogListeners[host] = listener
}

func (mux *LoggedMux) RemoveLogListener(host string) {
	delete(mux.VhostLogListeners, host)
}

type LogRecord struct {
	http.ResponseWriter
	status        int
	responseBytes int64
}

func (r *LogRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *LogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (mux *LoggedMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	record := &LogRecord{
		ResponseWriter: w,
	}

	// serve request and capture timings
	startTime := time.Now()
	mux.ServeMux.ServeHTTP(record, r)
	finishTime := time.Now()
	elapsedTime := finishTime.Sub(startTime)
	host := getHostName(r.Host)

	l := fmt.Sprintf("%s [%s] %s [ %d ] %s %d %s", r.RemoteAddr, host, r.Method, record.status, r.URL, r.ContentLength, elapsedTime)
	log.Println(l)

	if listener, ok := mux.VhostLogListeners[host]; ok {
		listener <- l
	}
}

// ignore port num, if any
func getHostName(input string) string {
	s := strings.Split(input, ":")
	return s[0]
}
