package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

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

		dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
		if err != nil {
			return fmt.Errorf("unable to create connection pool: %w", err)
		}
		defer dbpool.Close()

		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		r.GET("/xtz/delegations", api.GetDelegations(dbpool))

		log.Println("Server starting on :8080")
		if err := r.Run(":8080"); err != nil {
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
