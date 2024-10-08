package handlers

import (
	"HTTP-download-server/server/models"
	"net/http"
	"testing"

	"github.com/restsend/carrot"
	"github.com/stretchr/testify/assert"
)

func TestGetTaskList(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)

	db.Create(&models.Task{
		ID:       "task1",
		Name:     "Test Task 1",
		Status:   models.TaskFilterDownloaded,
		Size:     100,
		Url:      "http://example.com/file1",
		ChunkNum: 1,
	})
	db.Create(&models.Task{
		ID:       "task2",
		Name:     "Test Task 2",
		Status:   models.TaskStatusDownloading,
		Size:     200,
		Url:      "http://example.com/file2",
		ChunkNum: 2,
	})

	t.Run("successfulGetTasks", func(t *testing.T) {
		req := []byte(`{"status":"downloaded"}`)
		w := c.Post("POST", "/api/task/list", req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"totalCount":1`)
		assert.Contains(t, w.Body.String(), `"name":"Test Task 1"`)

		req = []byte(`{"status":"all"}`)
		w = c.Post("POST", "/api/task/list", req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"totalCount":2`)
	})

	t.Run("inputError", func(t *testing.T) {
		req := []byte(`{"invalid_json":}`)
		w := c.Post("POST", "/api/task/list", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDelete(t *testing.T) {
	r, db := createTestHandlers()
	c := carrot.NewTestClient(r)
	db.Create(&models.Task{
		ID:       "task1",
		Name:     "Test Task 1",
		Status:   models.TaskFilterDownloaded,
		Size:     100,
		Url:      "http://example.com/file1",
		ChunkNum: 1,
	})
	db.Create(&models.Task{
		ID:       "task2",
		Name:     "Test Task 2",
		Status:   models.TaskStatusDownloading,
		Size:     200,
		Url:      "http://example.com/file2",
		ChunkNum: 2,
	})
	t.Run("successfulDelete", func(t *testing.T) {
		req := []byte(`["task1", "task2"]`)
		w := c.Post("POST", "/api/task/delete", req)
		assert.Equal(t, http.StatusOK, w.Code)
		
		var task models.Task
		db.Where("id = ?", "task1").First(&task)
		assert.Equal(t, "", task.ID)
		db.Where("id = ?", "task2").First(&task)
		assert.Equal(t, "", task.ID)
	})

	t.Run("inputError", func(t *testing.T) {
		req := []byte(`{"invalid_json":}`)
		w := c.Post("POST", "/api/task/delete", req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
