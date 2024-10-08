package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
)

type EventSource struct {
	lastTime  time.Time
	key       string
	eventChan chan any
	ctx       context.Context
	cancel    context.CancelFunc
}

type EventSourceResult struct {
	Key string `json:"key"`
}

func (es *EventSource) Emit(event any) {
	if es.eventChan == nil {
		return
	}
	select {
	case es.eventChan <- event:
	default:
	}
}

func (h *Handlers) handleSSE(c *gin.Context) {
	c.Header("X-Accel-Buffering", "no")
	key := c.Param("key")
	v, ok := h.eventSources.Load(key)
	if !ok {
		c.JSON(http.StatusBadRequest, "missing key")
		return
	}

	eventSource := v.(*EventSource)

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			c.SSEvent("close", "user cancel download")
			return false
		case data, ok := <-eventSource.eventChan:
			if data == nil || !ok {
				c.SSEvent("close", "channel closed")
				return false
			}
			eventSource.lastTime = time.Now()
			byteData, _ := json.Marshal(data)
			c.SSEvent("message", string(byteData))
			return true
		}
	})
}

func (h *Handlers) createEventSourceInternal(key string) *EventSource {
	ctx, cancel := context.WithCancel(context.Background())

	eventSource := &EventSource{
		lastTime:  time.Now(),
		cancel:    cancel,
		ctx:       ctx,
		eventChan: make(chan any, 10),
		key:       key,
	}
	h.eventSources.Store(key, eventSource)

	go func() {
		defer h.cleanEventSource(key)
		for {
			select {
			case <-ctx.Done():
				carrot.Info("user cancel download", "key:", key)
				return
			case <-time.After(1 * time.Minute):
				if time.Since(eventSource.lastTime) > 30*time.Minute {
					carrot.Info("sse timeout", "key:", key)
					return
				}
			}
		}
	}()

	return eventSource
}

func (h *Handlers) createEventSource() *EventSource {
	key := carrot.RandText(8)
	return h.createEventSourceInternal(key)
}

func (h *Handlers) createEventSourceWithKey(key string) *EventSource {
	return h.createEventSourceInternal(key)
}

func (h *Handlers) cleanEventSource(key string) {
	v, ok := h.eventSources.LoadAndDelete(key)
	if !ok {
		return
	}

	eventSource, ok := v.(*EventSource)
	if !ok {
		return
	}

	eventSource.cancel()
	if eventSource.eventChan != nil {
		close(eventSource.eventChan)
		eventSource.eventChan = nil
	}
}
