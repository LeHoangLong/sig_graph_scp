package utility

import (
	"errors"
	"net/http"
	"sig_graph_scp/pkg/utility"

	"github.com/gin-gonic/gin"
)

func AbortWithError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, utility.ErrNotFound) {
		status = http.StatusNotFound
	} else if errors.Is(err, utility.ErrInvalidArgument) {
		status = http.StatusBadRequest
	}

	c.AbortWithStatusJSON(
		status,
		gin.H{"message": err.Error()},
	)
}

func AbortBadRequest(c *gin.Context, err error) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		gin.H{"message": err.Error()},
	)
}
