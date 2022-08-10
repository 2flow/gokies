package httputils

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-kit/log"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"time"
)

type HTTPServer struct {
	certManager *autocert.Manager
	routes      http.Handler
	logger      log.Logger
}

func NewHTTPServer(logger log.Logger) *HTTPServer {
	server := &HTTPServer{nil, nil, logger}
	return server
}

func (server *HTTPServer) configureAutoCert(domainURL string) {
	dataDir := "./cert"

	hostPolicy := func(ctx context.Context, host string) error {
		allowedHost := domainURL
		if host == allowedHost {
			return nil
		}
		return fmt.Errorf("acme/autocert: only %s host is allowed", allowedHost)
	}

	server.certManager = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache(dataDir),
	}
}

func (server *HTTPServer) SetRoutes(routes http.Handler) {
	server.routes = routes
}

func (server *HTTPServer) createHTTPServer() *http.Server {
	return &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func (server *HTTPServer) StartHTTPSServer() {
	httpServer := server.createHTTPServer()
	{
		httpServer.Handler = server.routes
		httpServer.Addr = ":443"
		httpServer.TLSConfig = &tls.Config{GetCertificate: server.certManager.GetCertificate}
	}

	go func() {
		err := httpServer.ListenAndServeTLS("", "")
		if err != nil {
			server.log("")
		}
	}()
}

func (server *HTTPServer) log(message string) {
	_ = server.logger.Log("[HTTPS Server]", message)
}

func (server *HTTPServer) StartHTTPServer() {
	httpServer := server.createHTTPServer()
	{
		if server.certManager != nil {
			httpServer.Handler = server.certManager.HTTPHandler(server.routes)
		} else {
			httpServer.Handler = server.routes
		}
		httpServer.Addr = ":80"
	}

	if err := httpServer.ListenAndServe(); err != nil {
		server.log(fmt.Sprintf("httpSrv.ListenAndServe() failed with %s", err))
	}
}
