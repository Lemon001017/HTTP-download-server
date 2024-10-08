package handlers

import (
	"HTTP-download-server/server/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

type FilterRequest struct {
	Status string `json:"status"`
}

func (h *Handlers) handleGetTaskList(c *gin.Context) {
	var filter FilterRequest
	err := c.ShouldBindJSON(&filter)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	tasks := models.GetTasksByStatus(h.db, filter.Status)

	resp := gin.H{
		"totalCount": len(tasks),
		"data":       tasks,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

func (h *Handlers) handleDelete(c *gin.Context) {
	var ids []string
	err := c.ShouldBindJSON(&ids)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	err = models.DeleteTasksByIds(h.db, ids)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}

	for _, id := range ids {
		models.DeleteChunks(h.db, id)
	}

	c.JSON(http.StatusOK, gin.H{"ids": ids})
}
