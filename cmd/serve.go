package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/broyeztony/delegated/internal/api"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long:  `Start the HTTP server to serve delegation data via REST API.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		log.Println("Serve command started")

		// Get database connection string
		connStr, err := getDatabaseURL()
		if err != nil {
			return err
		}

		// Initialize database connection
		dbpool, err := pgxpool.New(context.Background(), connStr)
		if err != nil {
			return fmt.Errorf("unable to create connection pool: %w", err)
		}
		defer dbpool.Close()

		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		r.GET("/xtz/delegations", api.GetDelegations(dbpool))

		// Create HTTP server with graceful shutdown
		server := &http.Server{
			Addr:    ":8080",
			Handler: r,
		}

		// Channel to listen for interrupt signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

		// Start server in a goroutine
		log.Println("Server starting on :8080")
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("server error: %v", err)
			}
		}()

		// Wait for shutdown signal
		<-sigChan
		log.Println("Shutdown signal received, gracefully shutting down...")

		// Create a context with 30-second timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown the server
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}

		log.Println("Server stopped")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
