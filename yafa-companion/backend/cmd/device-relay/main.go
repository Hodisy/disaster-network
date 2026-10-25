// Local Colima bridge only: the Mac LAN interface is outside the Docker VM.
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	address := os.Getenv("YAFA_DEVICE_BIND_ADDRESS")
	port := os.Getenv("YAFA_DEVICE_PORT")
	if port == "" {
		port = "8091"
	}
	ip := net.ParseIP(address)
	if ip == nil || !ip.IsPrivate() {
		log.Fatal("Set the Mac private LAN IPv4 address")
	}
	target, _ := url.Parse("http://127.0.0.1:" + port)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		http.Error(w, "Local device backend unavailable", http.StatusBadGateway)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/device/sync", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 2048)
		proxy.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: net.JoinHostPort(address, port), Handler: mux, ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 6 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 4096}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	log.Print("Yafa local LAN device relay starting; no admin routes")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Local device relay could not start")
	}
}
