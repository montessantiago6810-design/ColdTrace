// Command migrate applies or rolls back the ColdTrace database schema.
//
//	migrate up            apply every pending migration
//	migrate down-to N     roll back to version N (0 drops the schema)
//	migrate status        list migrations and whether they are applied
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/montessantiago6810-design/coldtrace/internal/platform/config"
	"github.com/montessantiago6810-design/coldtrace/internal/platform/logging"
	"github.com/montessantiago6810-design/coldtrace/internal/store"
)

const usage = "usage: migrate up | down-to <version> | status"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "coldtrace migrate:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New(usage)
	}

	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if err := cfg.RequireDatabase(); err != nil {
		return err
	}

	logger := logging.New(os.Stderr, cfg.LogFormat, cfg.LogLevel, "migrate")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := store.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	m, err := store.NewMigrator(pool, logger)
	if err != nil {
		return err
	}
	defer m.Close()

	switch {
	case args[0] == "up" && len(args) == 1:
		v, err := m.Up(ctx)
		if err != nil {
			return err
		}
		logger.Info("schema up to date", slog.Int64("version", v))
	case args[0] == "down-to" && len(args) == 2:
		target, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil || target < 0 {
			return fmt.Errorf("down-to: version must be a non-negative integer, got %q", args[1])
		}
		v, err := m.DownTo(ctx, target)
		if err != nil {
			return err
		}
		logger.Info("schema rolled back", slog.Int64("version", v))
	case args[0] == "status" && len(args) == 1:
		statuses, err := m.Status(ctx)
		if err != nil {
			return err
		}
		for _, s := range statuses {
			applied := "pending"
			if !s.AppliedAt.IsZero() {
				applied = "applied " + s.AppliedAt.UTC().Format("2006-01-02 15:04:05Z")
			}
			fmt.Printf("%05d  %-40s %s\n", s.Source.Version, s.Source.Path, applied)
		}
	default:
		return errors.New(usage)
	}
	return nil
}
