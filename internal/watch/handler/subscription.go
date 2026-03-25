package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yanking/price-watch/internal/watch/subscription"
)

type SubscriptionHandler struct {
	mgr *subscription.Manager
}

func NewSubscriptionHandler(mgr *subscription.Manager) *SubscriptionHandler {
	return &SubscriptionHandler{mgr: mgr}
}

func (h *SubscriptionHandler) Add(c *gin.Context) {
	var req subscription.SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Symbols) == 0 {
		Error(c, http.StatusBadRequest, "symbols is required")
		return
	}
	result, err := h.mgr.Add(c.Request.Context(), req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	OK(c, result)
}

func (h *SubscriptionHandler) Remove(c *gin.Context) {
	var req subscription.UnsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.mgr.Remove(c.Request.Context(), req)
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	OK(c, result)
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	infos, err := h.mgr.List(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	OK(c, infos)
}
