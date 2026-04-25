package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/favxlaw/mini-brimble/config"
	"github.com/favxlaw/mini-brimble/db"
	"github.com/favxlaw/mini-brimble/models"
	"github.com/favxlaw/mini-brimble/pipeline"
	"github.com/google/uuid"
)

type Handler struct {
	store       *db.DB
	broadcaster *pipeline.LogBroadcaster
	cfg         *config.Config
}

func New(store *db.DB, broadcaster *pipeline.LogBroadcaster, cfg *config.Config) *Handler {
	return &Handler{store: store, broadcaster: broadcaster, cfg: cfg}
}

func (h *Handler) CreateDeployment(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.SourceType == models.SourceTypeGit && req.SourceURL == "" {
		http.Error(w, "source_url is required for git deployments", http.StatusBadRequest)
		return
	}

	now := time.Now()
	deployment := &models.Deployment{
		ID:         uuid.New().String(),
		Name:       req.Name,
		SourceType: req.SourceType,
		Status:     models.StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if req.SourceURL != "" {
		deployment.SourceURL = &req.SourceURL
	}

	if err := h.store.CreateDeployment(deployment); err != nil {
		http.Error(w, "failed to create deployment", http.StatusInternalServerError)
		return
	}

	// Start the pipeline in the background.
	// We return 201 immediately — the client tracks progress via SSE.
	go pipeline.Deploy(h.store, h.broadcaster, h.cfg, deployment)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

func (h *Handler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	deployments, err := h.store.ListDeployments()
	if err != nil {
		http.Error(w, "failed to fetch deployments", http.StatusInternalServerError)
		return
	}

	// Return empty array not null when there are no deployments.
	// null breaks frontend code that calls .map() on the response.
	if deployments == nil {
		deployments = []*models.Deployment{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

func (h *Handler) GetDeployment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	deployment, err := h.store.GetDeployment(id)
	if err != nil {
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}
