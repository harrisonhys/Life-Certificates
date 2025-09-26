// @title Life Certificate Service API
// @version 1.0
// @description API for managing participants and life certificate verifications
// @BasePath /
// @securityDefinitions.basic BasicAuth
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "life-certificates/docs"
	"life-certificates/internal/config"
	"life-certificates/internal/database"
	"life-certificates/internal/frcore"
	httpserver "life-certificates/internal/http"
	"life-certificates/internal/http/handler"
	"life-certificates/internal/liveness"
	"life-certificates/internal/repository"
	"life-certificates/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.New(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	frClient, err := frcore.NewHTTPClient(frcore.Options{
		BaseURL:         cfg.FRC.BaseURL,
		UploadAPIKey:    cfg.FRC.UploadAPIKey,
		RecognizeAPIKey: cfg.FRC.RecognizeAPIKey,
		TenantID:        cfg.FRC.TenantID,
		Timeout:         cfg.FRC.RequestTimeout,
	})
	if err != nil {
		log.Fatalf("init fr client: %v", err)
	}

	participantRepo := repository.NewParticipantRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	certificateRepo := repository.NewLifeCertificateRepository(db)
	frIdentityRepo := repository.NewFRIdentityRepository(db)

	participantService := service.NewParticipantService(participantRepo, frIdentityRepo, certificateRepo, frClient)
	memberService := service.NewMemberService(memberRepo)
	checker := liveness.NoopChecker{Enabled: cfg.Liveness.Enabled}
	verificationService := service.NewVerificationService(participantRepo, certificateRepo, frIdentityRepo, frClient, checker, cfg.Verification.DistanceThreshold, cfg.Verification.SimilarityThreshold)

	participantHandler := handler.NewParticipantHandler(participantService)
	memberHandler := handler.NewMemberHandler(memberService)
	lifeHandler := handler.NewLifeCertificateHandler(verificationService)

	srv := httpserver.NewServer(cfg, participantHandler, memberHandler, lifeHandler)

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("HTTP server listening on %s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-sigCtx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	log.Println("server stopped cleanly")
}
