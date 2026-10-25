// The audio lab never initializes memory, database, bus, scheduler or agent services.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yafa/backend/connectors/audiohttp"
	connector "yafa/backend/connectors/services"
	"yafa/backend/core"
)

func main() {
	token := os.Getenv("YAFA_AUDIO_LAB_TOKEN")
	if len(token) < 32 {
		log.Fatal("Set a private YAFA_AUDIO_LAB_TOKEN of at least 32 characters")
	}
	provider := connector.OpenAIAudioLab{Key: os.Getenv("OPENAI_API_KEY"), TranscriptModel: os.Getenv("YAFA_TRANSCRIBE_MODEL"), FFmpeg: os.Getenv("YAFA_FFMPEG"), Client: &http.Client{Timeout: 60 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	model, voice := os.Getenv("OPENAI_LIVE_MODEL"), os.Getenv("OPENAI_LIVE_VOICE")
	if model == "" {
		model = "gpt-live-1"
	}
	if voice == "" {
		voice = "marin"
	}
	key := os.Getenv("OPENAI_API_KEY")
	if os.Getenv("YAFA_AUDIO_LAB_OFFLINE") == "1" {
		key = ""
		provider.Key = ""
	}
	live := core.NewAudioLabLive(ctx, connector.OpenAILive{APIKey: key, Model: model}, model, voice)
	live.BackendModel = os.Getenv("OPENAI_LIVE_BACKEND_MODEL")
	if live.BackendModel == "" {
		live.BackendModel = "gpt-5.6-luna"
	}
	handler := audiohttp.Handler{Live: live, Core: core.AudioLab{Service: provider}, Token: token, TranscriptionConfigured: provider.Key != ""}
	server := &http.Server{Addr: "127.0.0.1:4181", Handler: handler.Routes(), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 75 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8192}
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
		live.Shutdown(shutdown)
	}()
	log.Print("Yafa audio lab ready at http://127.0.0.1:4181; USB audio only, no database or bus")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Audio lab listener unavailable")
	}
	<-shutdownDone
}
