package api

import (
	"net/http"

	"github.com/OpenALC/LiteTalk/server/internal/store"
)

type createChannelReq struct {
	Name     string            `json:"name"`
	Type     store.ChannelType `json:"type"`
	Position *int              `json:"position"`
}

func validChannelName(s string) bool {
	n := len([]rune(s))
	return n >= 1 && n <= 64
}

// handleCreateChannel POST /api/v1/servers/{id}/channels（owner/admin）
func (s *Server) handleCreateChannel(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	serverID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if !s.requireRole(w, r, serverID, c.UserID, store.RoleAdmin) {
		return
	}
	var req createChannelReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if !validChannelName(req.Name) {
		writeErr(w, http.StatusBadRequest, "bad_name", "频道名需为 1–64 个字符")
		return
	}
	if req.Type != store.ChannelText && req.Type != store.ChannelVoice {
		writeErr(w, http.StatusBadRequest, "bad_type", "类型必须是 text 或 voice")
		return
	}
	ch := &store.Channel{
		ServerID: serverID,
		Name:     req.Name,
		Type:     req.Type,
	}
	if req.Position != nil {
		ch.Position = *req.Position
	}
	if err := s.store.CreateChannel(r.Context(), ch); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ch)
}

type patchChannelReq struct {
	Name     *string            `json:"name"`
	Type     *store.ChannelType `json:"type"`
	Position *int               `json:"position"`
}

// handlePatchChannel PATCH /api/v1/channels/{id}（owner/admin）
func (s *Server) handlePatchChannel(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	ch, err := s.store.ChannelByID(r.Context(), id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if !s.requireRole(w, r, ch.ServerID, c.UserID, store.RoleAdmin) {
		return
	}
	var req patchChannelReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name != nil {
		if !validChannelName(*req.Name) {
			writeErr(w, http.StatusBadRequest, "bad_name", "频道名需为 1–64 个字符")
			return
		}
		ch.Name = *req.Name
	}
	if req.Type != nil {
		if *req.Type != store.ChannelText && *req.Type != store.ChannelVoice {
			writeErr(w, http.StatusBadRequest, "bad_type", "类型必须是 text 或 voice")
			return
		}
		ch.Type = *req.Type
	}
	if req.Position != nil {
		ch.Position = *req.Position
	}
	if err := s.store.UpdateChannel(r.Context(), ch); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ch)
}

// handleDeleteChannel DELETE /api/v1/channels/{id}（owner/admin）
func (s *Server) handleDeleteChannel(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	ch, err := s.store.ChannelByID(r.Context(), id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if !s.requireRole(w, r, ch.ServerID, c.UserID, store.RoleAdmin) {
		return
	}
	if err := s.store.DeleteChannel(r.Context(), id); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
