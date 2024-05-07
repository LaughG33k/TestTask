package iternal

import (
	"crypto/tls"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

type HttpServer struct {
	config HttpServerConfig
	server *http.Server
}

type HttpServerConfig struct {
	Host string
	Port string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Idletimeout  time.Duration

	TLSConfig *tls.Config

	MaxHandlers          int
	MaxConcurrentStreams uint32

	Hanler http.Handler
}

func HttpServerInit(config HttpServerConfig) (*HttpServer, error) {

	srv := &http.Server{
		Addr:         config.Host + ":" + config.Port,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.Idletimeout,
		TLSConfig:    config.TLSConfig,
		Handler:      config.Hanler,
	}

	srv2 := &http2.Server{
		MaxHandlers:          config.MaxHandlers,
		MaxConcurrentStreams: config.MaxConcurrentStreams,
	}

	if err := http2.ConfigureServer(srv, srv2); err != nil {
		return nil, err
	}

	return &HttpServer{
		config: config,
		server: srv,
	}, nil

}

func (s *HttpServer) StartServer() error {

	return s.server.ListenAndServe()

}

func (s *HttpServer) StartTlsServer(certFile string, keyfile string) error {

	return s.server.ListenAndServeTLS(certFile, keyfile)

}

func (s *HttpServer) Close() error {
	return s.server.Close()
}
