package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"example.com/field-dispatch-room/internal/dispatch"
	realtimeclient "example.com/field-dispatch-room/internal/infrai"
)

type server struct {
	client *realtimeclient.Client
}

type dispatchRequest struct {
	WorkOrderID string                `json:"work_order_id"`
	Photo       workOrderPhoto        `json:"photo"`
	Technicians []dispatch.Technician `json:"technicians"`
}

type workOrderPhoto struct {
	ObjectKey string `json:"object_key"`
	Caption   string `json:"caption"`
}

type roomState struct {
	WorkOrderID string         `json:"work_order_id"`
	Photo       workOrderPhoto `json:"photo"`
	Status      string         `json:"status"`
	Technician  string         `json:"technician_id"`
	FollowUp    string         `json:"follow_up"`
	UpdatedAt   string         `json:"updated_at"`
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}

	srv := &server{client: realtimeclient.New(key)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /dispatch", srv.createDispatch)
	log.Println("dispatch service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s *server) createDispatch(w http.ResponseWriter, r *http.Request) {
	var input dispatchRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || input.WorkOrderID == "" || input.Photo.ObjectKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "work_order_id, photo.object_key, and technicians are required"})
		return
	}

	technician, err := dispatch.Match(input.Technicians)
	if errors.Is(err, dispatch.ErrNoTechnician) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	channel := "work-order-" + safeID(input.WorkOrderID)
	state := roomState{
		WorkOrderID: input.WorkOrderID,
		Photo:       input.Photo,
		Status:      "dispatched",
		Technician:  technician.ID,
		FollowUp:    "inspect_on_arrival",
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	if err := s.client.CreateChannel(ctx, channel, stableKey("channel", input.WorkOrderID)); err != nil {
		s.writeUpstreamError(w, err)
		return
	}
	if err := s.client.Publish(ctx, channel, "dispatch.assigned", state, technician.ID, stableKey("state", input.WorkOrderID)); err != nil {
		s.writeUpstreamError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"channel": channel, "state": state})
}

func (s *server) writeUpstreamError(w http.ResponseWriter, err error) {
	var apiErr *realtimeclient.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Message, "code": apiErr.Code})
		return
	}
	log.Printf("dispatch publish: %v", err)
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "dispatch state could not be published"})
}

func safeID(value string) string {
	clean := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return '-'
	}, value)
	return strings.Trim(clean, "-")
}

func stableKey(operation, workOrderID string) string {
	sum := sha256.Sum256([]byte(operation + ":" + workOrderID))
	return hex.EncodeToString(sum[:])
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
