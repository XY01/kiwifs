package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kiwifs/kiwifs/internal/claims"
	"github.com/labstack/echo/v4"
)

const (
	minLeaseDuration = 1 * time.Minute
	maxLeaseDuration = 24 * time.Hour
)

func (h *Handlers) ClaimTask(c echo.Context) error {
	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Actor header required")
	}
	if h.claimStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "claims not enabled")
	}

	var body struct {
		Path          string `json:"path"`
		LeaseDuration string `json:"lease_duration"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if body.Path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}

	ctx := c.Request().Context()

	if _, err := h.store.Stat(ctx, body.Path); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}

	lease := 30 * time.Minute
	if body.LeaseDuration != "" {
		d, err := time.ParseDuration(body.LeaseDuration)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid lease_duration")
		}
		lease = d
	}
	if lease < minLeaseDuration {
		lease = minLeaseDuration
	}
	if lease > maxLeaseDuration {
		return echo.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("lease_duration must be <= %s", maxLeaseDuration))
	}

	claim, err := h.claimStore.Claim(ctx, body.Path, actor, lease)
	if err != nil {
		if errors.Is(err, claims.ErrAlreadyClaimed) {
			existing, _ := h.claimStore.ActiveClaim(ctx, body.Path)
			return c.JSON(http.StatusConflict, map[string]any{
				"error":        "already claimed",
				"active_claim": existing,
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, claim)
}

func (h *Handlers) ReleaseTask(c echo.Context) error {
	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Actor header required")
	}
	if h.claimStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "claims not enabled")
	}

	var body struct {
		Path string `json:"path"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if body.Path == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "path is required")
	}

	ctx := c.Request().Context()
	if err := h.claimStore.Release(ctx, body.Path, actor); err != nil {
		if errors.Is(err, claims.ErrNotHolder) {
			return echo.NewHTTPError(http.StatusForbidden, "not the current claim holder")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "released"})
}

func (h *Handlers) ListClaims(c echo.Context) error {
	if h.claimStore == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "claims not enabled")
	}
	ctx := c.Request().Context()
	active, err := h.claimStore.ListActive(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"claims": active})
}
