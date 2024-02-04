package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/mayankshah1607/task-executor/internal"
	"github.com/mayankshah1607/task-executor/server"
	"github.com/mayankshah1607/task-executor/service"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

var (
	numWorkers int
	queueSize  int
	host       string
	port       string
)

func main() {
	// Parse command line flags.
	flag.IntVar(&numWorkers, "n", 3, "Number of concurrent worker goroutines")
	flag.IntVar(&queueSize, "q", 16, "Maximum number of tasks that can be added to the queue")
	flag.StringVar(&host, "host", "0.0.0.0", "Host address on which the server runs")
	flag.StringVar(&port, "port", "8080", "Port on which server listens for connections")
	flag.Parse()

	// For graceful cleanup.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the task executor.
	taskExecutor := internal.NewExecutor(queueSize)
	taskExecutor.Spawn(numWorkers)
	defer taskExecutor.Stop()

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Start server.
	service := service.NewService(logger, taskExecutor)
	server := server.NewServer(server.ServerOpts{
		Service: service,
		Host:    host,
		Port:    port,
		Logger:  logger,
	})

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return service.Run(ctx) })
	g.Go(func() error {
		server.Run(ctx)
		return nil
	})
	if err := g.Wait(); err != nil {
		logger.Fatal().Err(err)
	}
}
