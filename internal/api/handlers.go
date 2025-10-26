package api

import (
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

func GetDelegations(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(c.Request.Context(),
			"SELECT id, delegator, timestamp, amount, level FROM delegations ORDER BY id DESC")
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
