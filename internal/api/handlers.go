package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/broyeztony/delegated/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DelegationResponse struct {
	Timestamp string `json:"timestamp"`
	Amount    string `json:"amount"`
	Delegator string `json:"delegator"`
	Level     string `json:"level"`
}

// validateYear validates the year parameter
func validateYear(yearParam string) (int, error) {
	if yearParam == "" {
		return 0, nil // No year filter
	}

	year, err := strconv.Atoi(yearParam)
	if err != nil {
		return 0, fmt.Errorf("year must be a number")
	}

	if year < 2000 || year > 2100 {
		return 0, fmt.Errorf("year must be between 2000-2100")
	}

	return year, nil
}

func GetDelegations(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get optional year parameter
		yearParam := c.Query("year")

		query := "SELECT id, delegator, timestamp, amount, level FROM delegations"
		args := []interface{}{}

		if yearParam != "" {
			year, err := validateYear(yearParam)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			query += " WHERE EXTRACT(YEAR FROM timestamp) = $1"
			args = append(args, year)
		}

		query += " ORDER BY timestamp DESC"

		rows, err := db.Query(c.Request.Context(), query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		delegations, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Delegation])
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		responseData := make([]DelegationResponse, 0, len(delegations))
		for _, d := range delegations {
			responseData = append(responseData, DelegationResponse{
				Timestamp: d.Timestamp.Format(time.RFC3339),
				Amount:    strconv.FormatInt(d.Amount, 10),
				Delegator: d.Delegator,
				Level:     strconv.FormatInt(int64(d.Level), 10),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"data": responseData,
		})
	}
}
