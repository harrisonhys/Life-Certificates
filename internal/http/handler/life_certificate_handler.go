package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"life-certificates/internal/http/response"
	"life-certificates/internal/service"
)

// LifeCertificateHandler exposes endpoints for verification and status queries.
type LifeCertificateHandler struct {
	service *service.VerificationService
}

// NewLifeCertificateHandler wires dependencies for life certificate endpoints.
func NewLifeCertificateHandler(service *service.VerificationService) *LifeCertificateHandler {
	return &LifeCertificateHandler{service: service}
}

// Verify godoc
// @Summary Submit life certificate verification
// @Tags LifeCertificate
// @Security BasicAuth
// @Accept multipart/form-data
// @Produce json
// @Param participant_id formData string true "Participant ID"
// @Param image formData file true "Selfie image"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /life-certificate/verify [post]
func (h *LifeCertificateHandler) Verify(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	participantID := r.FormValue("participant_id")
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

	out, err := h.service.Verify(r.Context(), service.VerifyInput{
		ParticipantID:    participantID,
		ImageBytes:       imageBytes,
		OriginalFilename: header.Filename,
	})
	if err != nil {
		switch err {
		case service.ErrParticipantNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"participant_id":      out.ParticipantID,
		"verification_status": string(out.Status),
		"similarity":          out.Similarity,
		"distance":            out.Distance,
		"verified_at":         out.VerifiedAt,
	})
}

// LatestStatus godoc
// @Summary Get latest life certificate status
// @Tags LifeCertificate
// @Security BasicAuth
// @Produce json
// @Param participant_id path string true "Participant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /life-certificate/status/{participant_id} [get]
func (h *LifeCertificateHandler) LatestStatus(w http.ResponseWriter, r *http.Request) {
	participantID := chi.URLParam(r, "participant_id")

	out, err := h.service.LatestStatus(r.Context(), participantID)
	if err != nil {
		switch err {
		case service.ErrParticipantNotFound:
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	var lastStatus interface{}
	if out.Status != "" {
		lastStatus = out.Status
	}

	data := map[string]interface{}{
		"participant_id": out.ParticipantID,
		"last_status":    lastStatus,
		"similarity":     out.Similarity,
		"distance":       out.Distance,
	}
	if out.VerifiedAt != nil {
		data["verified_at"] = out.VerifiedAt
	}

	response.Success(w, http.StatusOK, data)
}
