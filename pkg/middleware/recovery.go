package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"timesheet-api/internal/resp"
)

func RecoveryJSON() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		resp.Internal(c, "Internal server error")
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
