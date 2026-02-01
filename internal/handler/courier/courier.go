package courier

import (
	"errors"
	"net/http"
	"strconv"

	handlerErrors "github.com/cdxy1/go-courier-service/internal/handler/errors"
	"github.com/cdxy1/go-courier-service/internal/model"
	repo "github.com/cdxy1/go-courier-service/internal/repository/courier"
	usecase "github.com/cdxy1/go-courier-service/internal/usecase/courier"
	"github.com/labstack/echo/v4"
)

type CourierHandler struct {
	uc courierUsecase
}

func NewCourierHandler(uc courierUsecase) *CourierHandler {
	return &CourierHandler{uc: uc}
}

func (h *CourierHandler) GetByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	result, err := h.uc.GetOneById(c.Request().Context(), id)
	if err != nil {
		switch err {
		case usecase.ErrInvalidID:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		case repo.ErrCourierNotFound:
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			if errors.Is(err, usecase.ErrInvalidID) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			if errors.Is(err, repo.ErrCourierNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}

	response := &courierResponse{
		ID:            result.ID,
		Name:          result.Name,
		Phone:         result.Phone,
		Status:        result.Status,
		TransportType: result.TransportType,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *CourierHandler) GetAll(c echo.Context) error {
	result, err := h.uc.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	response := make([]*courierResponse, 0, len(result))
	for _, v := range result {
		courier := &courierResponse{
			ID:            v.ID,
			Name:          v.Name,
			Phone:         v.Phone,
			Status:        v.Status,
			TransportType: v.TransportType,
		}
		response = append(response, courier)
	}

	return c.JSON(http.StatusOK, response)
}

func (h *CourierHandler) Create(c echo.Context) error {
	var req createCourierRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": handlerErrors.ErrBadRequest.Error()})
	}

	courier := &model.CourierModel{
		Name:          req.Name,
		Phone:         req.Phone,
		Status:        req.Status,
		TransportType: req.TransportType,
	}

	id, err := h.uc.Create(c.Request().Context(), courier)
	if err != nil {
		switch err {
		case usecase.ErrInvalidName, usecase.ErrInvalidPhone:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		case repo.ErrPhoneExists:
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			if errors.Is(err, usecase.ErrInvalidName) || errors.Is(err, usecase.ErrInvalidPhone) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			if errors.Is(err, repo.ErrPhoneExists) {
				return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}

	return c.JSON(http.StatusCreated, map[string]int{"id": id})
}

func (h *CourierHandler) Update(c echo.Context) error {
	var req updateCourierRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": handlerErrors.ErrBadRequest.Error()})
	}

	courier := &model.CourierModel{
		ID:            req.ID,
		Name:          req.Name,
		Phone:         req.Phone,
		Status:        req.Status,
		TransportType: req.TransportType,
	}

	err := h.uc.Update(c.Request().Context(), courier)
	if err != nil {
		switch err {
		case usecase.ErrInvalidID, usecase.ErrInvalidName, usecase.ErrInvalidPhone:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		case repo.ErrCourierNotFound:
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		case repo.ErrPhoneExists:
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			if errors.Is(err, usecase.ErrInvalidID) || errors.Is(err, usecase.ErrInvalidName) || errors.Is(err, usecase.ErrInvalidPhone) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			if errors.Is(err, repo.ErrCourierNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
			}
			if errors.Is(err, repo.ErrPhoneExists) {
				return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
}
