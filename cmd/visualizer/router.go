package main

import (
	"net/http"
	"os"
	"os/signal"
	"context"
	"syscall"
	"log"
	"time"

	"visualizer/src/hmap"
	"visualizer/src/ws"

	"github.com/fatih/color"
)

func vizual(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonBytes := hmap.GetBucketsJSON(m, "buckets")
	w.Write(jsonBytes)
}

func vizual_old(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonBytes := hmap.GetBucketsJSON(m, "oldbuckets")
	w.Write(jsonBytes)
}

func vizual_hmap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res, _ := hmap.GetHmapJSON(hmap.GetHmap(m))
	w.Write(res)
}

func startServer(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/vizual", vizual)
	mux.HandleFunc("/vizual_old", vizual_old)
	mux.HandleFunc("/hmap", vizual_hmap)
	mux.Handle("/", http.FileServer(http.Dir("frontend/dist")))
	mux.HandleFunc("/ws", ws.Handler)

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	green := color.New(color.FgGreen)
	magenta := color.New(color.FgMagenta)

	go func() {
		green.Printf("Сервер запущен: ")
		magenta.Printf("http://localhost%s\n", port)
		magenta.Printf("http://localhost%s/vizual     --> buckets JSON\n", port)
		magenta.Printf("http://localhost%s/vizual_old --> old buckets JSON\n", port)
		magenta.Printf("http://localhost%s/hmap       --> hmap JSON\n\n", port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return srv.Shutdown(ctx)
}

