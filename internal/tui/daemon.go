package tui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/edorguez/football-wizard/internal/scheduler"
	"go.uber.org/zap"
)

func RunDaemon(sched *scheduler.Scheduler, log *zap.Logger) error {
	log.Info("starting daemon mode")

	err := sched.Start()
	if err != nil {
		return fmt.Errorf("starting scheduler: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Info("received signal, shutting down", zap.String("signal", sig.String()))

	sched.Stop()
	log.Info("daemon stopped")

	return nil
}
