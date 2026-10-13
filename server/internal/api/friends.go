package api

import (
	"net/http"
	"strings"

	"github.com/OpenALC/LiteTalk/server/internal/store"
)

// friendView 好友条目视图（内联对端用户信息 + 请求数 id）。
type friendView struct {
	ID       int64                `json:"id"`
	Status   store.FriendshipStatus `json:"status"`
	Incoming bool                 `json:"incoming"` // true = 对端发起（我是被请求方）
	User     *store.User          `json:"user"`
}

// handleListFriends GET /api/v1/friends
func (s *Server) handleListFriends(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	fs, err := s.store.FriendshipsOfUser(r.Context(), c.UserID)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	out := []friendView{}
	for _, f := range fs {
		other := f.FriendID
		incoming := false
		if f.FriendID == c.UserID {
			other = f.UserID
			incoming = true
		}
		u, err := s.store.UserByID(r.Context(), other)
		if err != nil {
			continue
		}
		out = append(out, friendView{ID: f.ID, Status: f.Status, Incoming: incoming, User: u})
	}
	writeJSON(w, http.StatusOK, out)
}

type friendRequestReq struct {
	Nickname string `json:"nickname"` // 昵称#tag 或单独字段
	Tag      string `json:"tag"`
}

// handleFriendRequest POST /api/v1/friends/requests（按 昵称#tag 添加）
func (s *Server) handleFriendRequest(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	var req friendRequestReq
	if !decodeJSON(w, r, &req) {
		return
	}
	nickname, tag := req.Nickname, req.Tag
	if strings.Contains(nickname, "#") {
		parts := strings.SplitN(nickname, "#", 2)
		nickname, tag = parts[0], strings.TrimSpace(parts[1])
	}
	if !validNickname(nickname) || len(tag) != 4 {
		writeErr(w, http.StatusBadRequest, "bad_target", "目标格式应为 昵称#4位tag")
		return
	}
	target, err := s.store.UserByNicknameTag(r.Context(), nickname, tag)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "用户不存在")
		return
	}
	if target.ID == c.UserID {
		writeErr(w, http.StatusBadRequest, "bad_target", "不能添加自己为好友")
		return
	}
	// 已存在任何方向的关系 → 冲突。
	if existing, err := s.store.FriendshipBetween(r.Context(), c.UserID, target.ID); err == nil {
		if existing.Status == store.FriendAccepted {
			writeErr(w, http.StatusConflict, "already_friends", "已经是好友")
		} else {
			writeErr(w, http.StatusConflict, "request_exists", "已存在好友请求")
		}
		return
	}
	f := &store.Friendship{
		UserID:   c.UserID,
		FriendID: target.ID,
		Status:   store.FriendPending,
	}
	if err := s.store.CreateFriendship(r.Context(), f); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, friendView{ID: f.ID, Status: f.Status, Incoming: false, User: target})
}

// handleFriendAccept POST /api/v1/friends/requests/{id}/accept
func (s *Server) handleFriendAccept(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	f, err := s.store.FriendshipByID(r.Context(), id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	// 只有被请求方可以接受。
	if f.FriendID != c.UserID {
		writeErr(w, http.StatusForbidden, "forbidden", "只有被请求方可以接受")
		return
	}
	if f.Status != store.FriendPending {
		writeErr(w, http.StatusConflict, "bad_state", "请求不是待处理状态")
		return
	}
	f.Status = store.FriendAccepted
	if err := s.store.UpdateFriendship(r.Context(), f); err != nil {
		writeStoreErr(w, err)
		return
	}
	other, _ := s.store.UserByID(r.Context(), f.UserID)
	writeJSON(w, http.StatusOK, friendView{ID: f.ID, Status: f.Status, Incoming: true, User: other})
}

// handleFriendDelete DELETE /api/v1/friends/{id}
func (s *Server) handleFriendDelete(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	f, err := s.store.FriendshipByID(r.Context(), id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if f.UserID != c.UserID && f.FriendID != c.UserID {
		writeErr(w, http.StatusForbidden, "forbidden", "无权操作")
		return
	}
	if err := s.store.DeleteFriendship(r.Context(), id); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
