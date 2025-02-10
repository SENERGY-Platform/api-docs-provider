package util

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
)

func NewServer(handler http.Handler, port int) *http.Server {
	return &http.Server{
		Addr:    ":" + strconv.FormatInt(int64(port), 10),
		Handler: handler,
	}
}

func StartServer(server *http.Server) error {
	Logger.Info("starting http server")
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func StopServer(ctx context.Context, server *http.Server) error {
	<-ctx.Done()
	Logger.Info("stopping http server")
	ctxWt, cf := context.WithTimeout(context.Background(), time.Second*5)
	defer cf()
	if err := server.Shutdown(ctxWt); err != nil {
		return err
	}
	Logger.Info("http server shutdown complete")
	return nil
}
