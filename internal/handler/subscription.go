package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"effective_mobile/internal/models"
	"effective_mobile/internal/service"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
)

// SubscriptionHandler wires service and logging
type SubscriptionHandler struct {
	svc      *service.SubscriptionService
	log      *slog.Logger
	validate *validator.Validate
}

func NewSubscriptionHandler(svc *service.SubscriptionService, log *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// @Summary      Создать подписку
// @Description  Добавляет новую подписку пользователю
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription body models.SubscriptionInput true "Subscription input"
// @Success      201 {object} models.Subscription
// @Failure      400 {object} map[string]string
// @Router       /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.SubscriptionInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("invalid json", slog.Any("err", err))
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.log.Error("validation failed", slog.Any("err", err))
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	parsedStart, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format", http.StatusBadRequest)
		return
	}

	var parsedEnd *time.Time
	if input.EndDate != nil {
		t, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
		parsedEnd = &t
	}

	sub := &models.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      uuid.UUID(input.UserID),
		StartDate:   parsedStart,
		EndDate:     parsedEnd,
	}

	if err := h.svc.Create(r.Context(), sub); err != nil {
		h.log.Error("failed to create subscription", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

// @Summary Получить подписку
// @Description Возвращает подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "ID подписки"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sub, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.log.Error("failed to get subscription", slog.Any("err", err))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

// @Summary Список подписок
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "UUID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param limit query int false "Лимит" default(10)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.Subscription
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var userID *uuid.UUID
	if val := query.Get("user_id"); val != "" {
		id, err := uuid.Parse(val)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		userID = &id
	}

	var serviceName *string
	if val := query.Get("service_name"); val != "" {
		serviceName = &val
	}

	limit := 10
	offset := 0
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := query.Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	list, err := h.svc.List(r.Context(), userID, serviceName, limit, offset)
	if err != nil {
		h.log.Error("list error", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}

// @Summary Обновить подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param subscription body models.SubscriptionInput true "Subscription"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} string
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var input models.SubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	parsedStart, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format", http.StatusBadRequest)
		return
	}

	var parsedEnd *time.Time
	if input.EndDate != nil {
		t, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
		parsedEnd = &t
	}

	sub := &models.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   parsedStart,
		EndDate:     parsedEnd,
	}

	if err := h.svc.Update(r.Context(), id, sub); err != nil {
		h.log.Error("update failed", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sub)
}

// @Summary Удалить подписку
// @Tags subscriptions
// @Param id path int true "ID"
// @Success 204
// @Failure 400
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.log.Error("delete failed", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Сумма подписок за период
// @Tags subscriptions
// @Produce json
// @Param user_id query string true "UUID"
// @Param service_name query string false "Фильтр по сервису"
// @Param from query string true "Формат MM-YYYY"
// @Param to query string true "Формат MM-YYYY"
// @Success 200 {object} map[string]int
// @Failure 400
// @Router /subscriptions/summary [get]
func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	userStr := q.Get("user_id")
	fromStr := q.Get("from")
	toStr := q.Get("to")

	if userStr == "" || fromStr == "" || toStr == "" {
		http.Error(w, "missing required params", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	from, err := time.Parse("01-2006", fromStr)
	if err != nil {
		http.Error(w, "invalid from date", http.StatusBadRequest)
		return
	}
	to, err := time.Parse("01-2006", toStr)
	if err != nil {
		http.Error(w, "invalid to date", http.StatusBadRequest)
		return
	}

	var serviceName *string
	if val := q.Get("service_name"); val != "" {
		serviceName = &val
	}

	total, err := h.svc.SumByPeriod(r.Context(), userID, serviceName, from, to)
	if err != nil {
		h.log.Error("summary error", slog.Any("err", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"total": total})
}

func RegisterRoutes(h *SubscriptionHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(h.LoggingMiddleware)

	r.Post("/subscriptions", h.Create)
	r.Get("/subscriptions", h.List)
	r.Get("/subscriptions/{id}", h.Get)
	r.Put("/subscriptions/{id}", h.Update)
	r.Delete("/subscriptions/{id}", h.Delete)
	r.Get("/subscriptions/summary", h.Summary)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	return r
}
