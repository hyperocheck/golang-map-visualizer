package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"visualizer/src/engine"
	"visualizer/src/ws"

	"github.com/fatih/color"
)

func startServer[K comparable, V any](t *engine.Meta[K, V], port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/vizual", t.VisualHandler)
	mux.HandleFunc("/vizual_old", t.VisualOldHandler)
	mux.HandleFunc("/hmap", t.HmapHandler)
	mux.HandleFunc("/delete_key", t.DeleteKey)
	mux.HandleFunc("/update_key", t.UpdateKey)
	mux.Handle("/", http.FileServer(http.Dir("frontend/dist")))
	mux.HandleFunc("/ws", ws.Handler)

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	green := color.New(color.FgGreen).SprintfFunc()

	go func() {
		t.Console.PrintlnLog(green("Go to http://localhost%s\n", port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
