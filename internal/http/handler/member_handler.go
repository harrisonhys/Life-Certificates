package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"life-certificates/internal/http/response"
	"life-certificates/internal/service"
)

// MemberHandler exposes member CRUD endpoints.
type MemberHandler struct {
	service *service.MemberService
}

// NewMemberHandler wires dependencies for member endpoints.
func NewMemberHandler(service *service.MemberService) *MemberHandler {
	return &MemberHandler{service: service}
}

// Create godoc
// @Summary Create member
// @Description Create a new member record
// @Tags Members
// @Security BasicAuth
// @Accept json
// @Produce json
// @Param payload body service.CreateMemberInput true "Member payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /members [post]
func (h *MemberHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateMemberInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	member, err := h.service.Create(r.Context(), req)
	if err != nil {
		switch err {
		case service.ErrMemberNIKExists, service.ErrMemberNomorPesertaExists:
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.Success(w, http.StatusCreated, member)
}

// List godoc
// @Summary List members
// @Tags Members
// @Security BasicAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /members [get]
func (h *MemberHandler) List(w http.ResponseWriter, r *http.Request) {
	members, err := h.service.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{"members": members})
}

// Get godoc
// @Summary Get member detail
// @Tags Members
// @Security BasicAuth
// @Produce json
// @Param member_id path string true "Member ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /members/{member_id} [get]
func (h *MemberHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "member_id")
	member, err := h.service.Get(r.Context(), id)
	if err != nil {
		switch err {
		case service.ErrMemberNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(w, http.StatusOK, member)
}

// Update godoc
// @Summary Update member data
// @Tags Members
// @Security BasicAuth
// @Accept json
// @Produce json
// @Param member_id path string true "Member ID"
// @Param payload body service.UpdateMemberInput true "Update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /members/{member_id} [put]
func (h *MemberHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "member_id")
	var req service.UpdateMemberInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	member, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		switch err {
		case service.ErrMemberNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		case service.ErrMemberNIKExists, service.ErrMemberNomorPesertaExists:
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.Success(w, http.StatusOK, member)
}

// Delete godoc
// @Summary Delete member
// @Tags Members
// @Security BasicAuth
// @Param member_id path string true "Member ID"
// @Success 204 {string} string ""
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /members/{member_id} [delete]
func (h *MemberHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "member_id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		switch err {
		case service.ErrMemberNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
