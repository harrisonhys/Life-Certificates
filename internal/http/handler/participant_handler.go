package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"life-certificates/internal/http/response"
	"life-certificates/internal/service"
)

// ParticipantHandler exposes participant related endpoints.
type ParticipantHandler struct {
	service *service.ParticipantService
}

// NewParticipantHandler wires dependencies for participant endpoints.
func NewParticipantHandler(service *service.ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{service: service}
}

// Register godoc
// @Summary Register participant
// @Description Register participant and store reference with FR Core
// @Tags Participants
// @Security BasicAuth
// @Accept multipart/form-data
// @Produce json
// @Param nik formData string true "Participant NIK"
// @Param name formData string true "Participant name"
// @Param image formData file true "Initial selfie image"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /participants/register [post]
func (h *ParticipantHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "failed to read image")
		return
	}

	out, err := h.service.Register(r.Context(), service.RegisterInput{
		NIK:       r.FormValue("nik"),
		Name:      r.FormValue("name"),
		Image:     imageBytes,
		ImageName: header.Filename,
	})
	if err != nil {
		switch err {
		case service.ErrParticipantExists:
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.Success(w, http.StatusCreated, map[string]interface{}{
		"participant_id":  out.ParticipantID,
		"fr_ref":          out.FRRef,
		"fr_external_ref": out.FRExternalRef,
	})
}

// List godoc
// @Summary List participants
// @Tags Participants
// @Security BasicAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /participants [get]
func (h *ParticipantHandler) List(w http.ResponseWriter, r *http.Request) {
	participants, err := h.service.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{"participants": participants})
}

// Get godoc
// @Summary Get participant detail
// @Tags Participants
// @Security BasicAuth
// @Produce json
// @Param participant_id path string true "Participant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /participants/{participant_id} [get]
func (h *ParticipantHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "participant_id")
	participant, err := h.service.Get(r.Context(), id)
	if err != nil {
		switch err {
		case service.ErrParticipantNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(w, http.StatusOK, participant)
}

// Update godoc
// @Summary Update participant metadata
// @Tags Participants
// @Security BasicAuth
// @Accept json
// @Produce json
// @Param participant_id path string true "Participant ID"
// @Param payload body service.UpdateParticipantInput true "Update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /participants/{participant_id} [put]
func (h *ParticipantHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "participant_id")
	var req service.UpdateParticipantInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	participant, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		switch err {
		case service.ErrParticipantNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		case service.ErrParticipantExists:
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(w, http.StatusOK, participant)
}

// Delete godoc
// @Summary Delete participant
// @Tags Participants
// @Security BasicAuth
// @Param participant_id path string true "Participant ID"
// @Success 204 {string} string ""
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /participants/{participant_id} [delete]
func (h *ParticipantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "participant_id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		switch err {
		case service.ErrParticipantNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
