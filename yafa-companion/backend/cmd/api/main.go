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
	"yafa/backend/connectors/config"
	"yafa/backend/connectors/httpapi"
	repositoryconnectors "yafa/backend/connectors/repositories"
	serviceconnectors "yafa/backend/connectors/services"
	"yafa/backend/core"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	repository := repositoryconnectors.MemoryHTTP{BaseURL: cfg.MemoryURL, Token: cfg.MemoryToken, Client: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	agentRepository := repositoryconnectors.AgentMemory{Memory: repository}
	bootstrap := core.WorkspaceCore{Repository: agentRepository, Now: time.Now}
	initialization, cancelInitialization := context.WithTimeout(ctx, 30*time.Second)
	err = bootstrap.InitializeContextFiles(initialization)
	cancelInitialization()
	if err != nil {
		// Do not log documents, provider configuration or raw connector errors.
		log.Fatal("Agent context initialization failed: check Strata availability, stored document validity and workspace conflicts; the API was not started")
	}
	var prices *serviceconnectors.TokenPrices
	if len(cfg.TokenPrices) == 3 {
		prices = &serviceconnectors.TokenPrices{Input: cfg.TokenPrices[0], CachedInput: cfg.TokenPrices[1], Output: cfg.TokenPrices[2]}
	}
	model := serviceconnectors.OpenAIResponses{APIKey: cfg.OpenAIKey, Model: cfg.OpenAIModel, Prices: prices, Client: &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}}
	agent := core.NewAgentHarness(ctx, agentRepository, model, cfg.AgentLimits)
	agent.Features = cfg.Features
	agent.Safety = &core.SafetyCore{Repository: repository, Limits: cfg.Safety}
	agent.Files = repositoryconnectors.FileMemory{AgentMemory: agentRepository}
	agent.Hosted = serviceconnectors.OpenAIHosted{APIKey: cfg.OpenAIKey, AnalysisModel: cfg.OpenAIModel, ImageModel: cfg.ImageModel, Prices: prices, Client: &http.Client{Timeout: 70 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}}
	codexRunner := serviceconnectors.CodexSSH{Host: cfg.CodexHost, User: cfg.CodexUser, Port: cfg.CodexPort, KeyFile: cfg.CodexKeyFile, KnownHostsFile: cfg.CodexKnownHostsFile, Runner: cfg.CodexRunner}
	agent.Codex = core.NewCodexJobs(ctx, repositoryconnectors.CodexMemory{Memory: repository}, codexRunner, agent.Safety, agent.RuntimeID, cfg.CodexMaxSeconds)
	voice := core.NewVoiceSessions(agent, serviceconnectors.OpenAILive{APIKey: cfg.OpenAIKey, Model: cfg.LiveModel}, cfg.LiveModel, cfg.LiveVoice, ctx, cfg.LiveMaxSeconds)
	voice.IdleSeconds = cfg.LiveIdleSeconds
	voice.USDPerMinute = cfg.LivePrice
	handler := httpapi.Handler{Memory: core.Memory{Repository: repository}, Token: cfg.APIToken, Device: core.DeviceLink{Repository: repositoryconnectors.DeviceLinkMemory{Memory: repository}, DeviceID: cfg.DeviceID, Now: time.Now}, DeviceToken: cfg.DeviceToken, Agent: agent, Voice: voice}
	server := &http.Server{Addr: cfg.Listen, Handler: handler.Routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	var deviceServer *http.Server
	if cfg.DeviceListen != "" {
		deviceServer = &http.Server{Addr: cfg.DeviceListen, Handler: handler.DeviceRoutes(), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 12 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 4096}
	}
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
		if deviceServer != nil {
			_ = deviceServer.Shutdown(shutdown)
		}
		agent.Shutdown(shutdown)
		voice.Shutdown(shutdown)
	}()
	if deviceServer != nil {
		go func() {
			if err := deviceServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Print("Device listener stopped unexpectedly")
				stop()
			}
		}()
	}
	log.Print("Yafa API starting; single-person infrastructure slice")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("HTTP server stopped unexpectedly")
	}
	<-shutdownDone
}
