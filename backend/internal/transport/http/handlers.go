package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/szabolcs/cms/internal/domain"
	"github.com/szabolcs/cms/internal/repository"
	"github.com/szabolcs/cms/internal/service"
)

// Handler holds all HTTP handler dependencies.
type Handler struct {
	validation *service.ValidationService
	vehicles   repository.VehicleRepository
	stops      repository.StopRepository
	odMatrix   repository.ODMatrixRepository
	stats      repository.StatsRepository
	events     repository.ValidationRepository
	logger     *slog.Logger
}

// NewHandler creates a new Handler with all dependencies.
func NewHandler(
	validation *service.ValidationService,
	vehicles repository.VehicleRepository,
	stops repository.StopRepository,
	odMatrix repository.ODMatrixRepository,
	stats repository.StatsRepository,
	events repository.ValidationRepository,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		validation: validation,
		vehicles:   vehicles,
		stops:      stops,
		odMatrix:   odMatrix,
		stats:      stats,
		events:     events,
		logger:     logger,
	}
}

// RegisterRoutes sets up all API routes on the given engine.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/checkin", h.checkin)
		v1.POST("/checkout", h.checkout)
		v1.GET("/od-matrix", h.getODMatrix)
		v1.GET("/vehicles", h.getVehicles)
		v1.GET("/stops", h.getStops)
		v1.GET("/events/recent", h.getRecentEvents)
		v1.GET("/stats", h.getStats)
	}
}

func (h *Handler) checkin(c *gin.Context) {
	var req service.CheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, CodeValidationError, err.Error())
		return
	}

	event, err := h.validation.Checkin(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respondOK(c, event)
}

func (h *Handler) checkout(c *gin.Context) {
	var req service.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, CodeValidationError, err.Error())
		return
	}

	event, err := h.validation.Checkout(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respondOK(c, event)
}

func (h *Handler) getODMatrix(c *gin.Context) {
	rows, err := h.odMatrix.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get OD matrix", "error", err)
		respondError(c, http.StatusInternalServerError, CodeInternalError, "failed to fetch OD matrix")
		return
	}
	respondOK(c, rows)
}

func (h *Handler) getVehicles(c *gin.Context) {
	vehicles, err := h.vehicles.FindAll(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get vehicles", "error", err)
		respondError(c, http.StatusInternalServerError, CodeInternalError, "failed to fetch vehicles")
		return
	}
	respondOK(c, vehicles)
}

func (h *Handler) getStops(c *gin.Context) {
	stops, err := h.stops.FindAll(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get stops", "error", err)
		respondError(c, http.StatusInternalServerError, CodeInternalError, "failed to fetch stops")
		return
	}
	respondOK(c, stops)
}

func (h *Handler) getRecentEvents(c *gin.Context) {
	events, err := h.events.RecentEvents(c.Request.Context(), 20)
	if err != nil {
		h.logger.Error("failed to get recent events", "error", err)
		respondError(c, http.StatusInternalServerError, CodeInternalError, "failed to fetch events")
		return
	}
	respondOK(c, events)
}

func (h *Handler) getStats(c *gin.Context) {
	s, err := h.stats.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get stats", "error", err)
		respondError(c, http.StatusInternalServerError, CodeInternalError, "failed to fetch stats")
		return
	}
	respondOK(c, s)
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		respondError(c, http.StatusNotFound, CodePassengerNotFound, "passenger not found")
	case errors.Is(err, domain.ErrPassengerInactive):
		respondError(c, http.StatusForbidden, CodePassengerInactive, "passenger is not active")
	case errors.Is(err, domain.ErrOpenCheckinExists):
		respondError(c, http.StatusConflict, CodeOpenCheckinExists, "open checkin already exists")
	default:
		h.logger.Error("service error", "error", err)
		respondError(c, http.StatusInternalServerError, CodeInternalError, "internal server error")
	}
}
