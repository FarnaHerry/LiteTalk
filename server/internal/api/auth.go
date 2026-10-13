package api

import (
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/OpenALC/LiteTalk/server/internal/auth"
	"github.com/OpenALC/LiteTalk/server/internal/store"
)

// ---- 请求/响应结构 ----

type registerReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

type emailCodeReq struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"` // register | reset
}

type passwordResetReq struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

// tokenPair 认证响应。
type tokenPair struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresIn    int64      `json:"expires_in"` // 秒
	User         *store.User `json:"user"`
}

// ---- 校验 ----

// validEmail 简单邮箱格式校验。
func validEmail(s string) bool {
	if len(s) == 0 || len(s) > 255 {
		return false
	}
	_, err := mail.ParseAddress(s)
	return err == nil
}

// validNickname 昵称 2–16 个字符。
func validNickname(s string) bool {
	n := len([]rune(s))
	return n >= 2 && n <= 16
}

// validPassword 密码 ≥8 位。
func validPassword(s string) bool { return len(s) >= 8 }

// fmtTag 定宽 4 位。
func fmtTag(n int) string {
	s := strconv.Itoa(n % 10000)
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

// parseRefreshUID 解析 "<uid>.<random>" 格式刷新令牌中的用户 ID。
func parseRefreshUID(token string) (int64, bool) {
	return auth.ParseRefreshUID(token)
}

// ---- handlers ----

// handleRegister POST /api/v1/auth/register
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = normalizeEmail(req.Email)
	if !validNickname(req.Nickname) {
		writeErr(w, http.StatusBadRequest, "bad_nickname", "昵称需为 2–16 个字符")
		return
	}
	if !validEmail(req.Email) {
		writeErr(w, http.StatusBadRequest, "bad_email", "邮箱格式不正确")
		return
	}
	if !validPassword(req.Password) {
		writeErr(w, http.StatusBadRequest, "bad_password", "密码至少 8 位")
		return
	}
	// 验证邮箱验证码。
	if err := s.codes.Verify(r.Context(), auth.PurposeRegister, req.Email, req.Code); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_code", "验证码无效或已过期")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeStoreErr(w, err)
		return
	}

	// 分配 tag：4 位随机，冲突重试（UNIQUE(nickname,tag)）。
	var user *store.User
	for i := 0; i < 5; i++ {
		user = &store.User{
			Email:         req.Email,
			Nickname:      req.Nickname,
			Tag:           fmtTag(int(time.Now().UnixNano()+int64(i*7919)) % 10000),
			PasswordHash:  hash,
			EmailVerified: true, // 邮箱已过验证码验证
		}
		err = s.store.CreateUser(r.Context(), user)
		if errors.Is(err, store.ErrConflict) {
			// 区分邮箱冲突与 tag 冲突。
			if _, e := s.store.UserByEmail(r.Context(), req.Email); e == nil {
				writeErr(w, http.StatusConflict, "email_taken", "邮箱已注册")
				return
			}
			continue // tag 冲突重试
		}
		break
	}
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	s.issueTokens(w, r, user, http.StatusCreated)
}

// handleLogin POST /api/v1/auth/login（限流 5 次/min/IP）
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ok, err := s.limits.Allow(r.Context(), "login:"+clientIP(r), 5, time.Minute)
	if err != nil || !ok {
		writeErr(w, http.StatusTooManyRequests, "rate_limited", "登录过于频繁，请稍后再试")
		return
	}
	var req loginReq
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = normalizeEmail(req.Email)
	user, err := s.store.UserByEmail(r.Context(), req.Email)
	if err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeErr(w, http.StatusUnauthorized, "bad_credentials", "邮箱或密码错误")
		s.log.Printf("[audit] login failed email=%s ip=%s", req.Email, clientIP(r))
		return
	}
	s.log.Printf("[audit] login ok user=%d ip=%s", user.ID, clientIP(r))
	s.issueTokens(w, r, user, http.StatusOK)
}

// handleRefresh POST /api/v1/auth/refresh（旋转刷新令牌）
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.RefreshToken == "" {
		writeErr(w, http.StatusBadRequest, "bad_request", "缺少 refresh_token")
		return
	}
	// 旧令牌哈希反查：逐用户扫不可行，refresh 存储按 userID 键控；
	// 客户端 refresh 请求附带 access_token 解析 uid 更稳——此处约定请求体可带 user_id，
	// 或由客户端用过期 access 中的 uid。简化：请求体带 user_id。
	// 为兼容只传 refresh_token 的客户端，token 中嵌入 uid 前缀。
	uid, ok := parseRefreshUID(req.RefreshToken)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized", "refresh_token 不合法")
		return
	}
	plain, err := s.refresh.Rotate(r.Context(), uid, req.RefreshToken)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "unauthorized", "refresh_token 无效或已吊销")
		return
	}
	user, err := s.store.UserByID(r.Context(), uid)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	access, err := s.jwt.SignAccess(user.ID, user.Nickname, user.Tag)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tokenPair{
		AccessToken:  access,
		RefreshToken: plain,
		ExpiresIn:    int64(auth.AccessTokenTTL / time.Second),
		User:         user,
	})
}

// handleLogout POST /api/v1/auth/logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req logoutReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if uid, ok := parseRefreshUID(req.RefreshToken); ok {
		_ = s.refresh.Revoke(r.Context(), uid)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleEmailCode POST /api/v1/auth/email-code
func (s *Server) handleEmailCode(w http.ResponseWriter, r *http.Request) {
	var req emailCodeReq
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = normalizeEmail(req.Email)
	if !validEmail(req.Email) {
		writeErr(w, http.StatusBadRequest, "bad_email", "邮箱格式不正确")
		return
	}
	if req.Purpose != auth.PurposeRegister && req.Purpose != auth.PurposeReset {
		writeErr(w, http.StatusBadRequest, "bad_purpose", "用途必须是 register 或 reset")
		return
	}
	// register：邮箱不应已注册；reset：邮箱必须已注册（防枚举差异统一提示）。
	_, lookupErr := s.store.UserByEmail(r.Context(), req.Email)
	if req.Purpose == auth.PurposeRegister && lookupErr == nil {
		writeErr(w, http.StatusConflict, "email_taken", "邮箱已注册")
		return
	}
	if req.Purpose == auth.PurposeReset && lookupErr != nil {
		writeErr(w, http.StatusNotFound, "not_found", "邮箱未注册")
		return
	}
	code, err := s.codes.Send(r.Context(), req.Purpose, req.Email)
	switch {
	case errors.Is(err, auth.ErrCodeTooFrequent):
		writeErr(w, http.StatusTooManyRequests, "too_frequent", "发送过于频繁，请 60 秒后再试")
		return
	case errors.Is(err, auth.ErrCodeDailyLimit):
		writeErr(w, http.StatusTooManyRequests, "daily_limit", "今日发送次数已达上限")
		return
	case err != nil:
		writeStoreErr(w, err)
		return
	}
	if err := s.mailer.SendCode(req.Email, req.Purpose, code); err != nil {
		s.log.Printf("[mailer] 发送失败 to=%s err=%v", req.Email, err)
	}
	s.log.Printf("[audit] email-code sent email=%s purpose=%s ip=%s", req.Email, req.Purpose, clientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handlePasswordReset POST /api/v1/auth/password/reset
func (s *Server) handlePasswordReset(w http.ResponseWriter, r *http.Request) {
	var req passwordResetReq
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = normalizeEmail(req.Email)
	if !validPassword(req.Password) {
		writeErr(w, http.StatusBadRequest, "bad_password", "密码至少 8 位")
		return
	}
	if err := s.codes.Verify(r.Context(), auth.PurposeReset, req.Email, req.Code); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_code", "验证码无效或已过期")
		return
	}
	user, err := s.store.UserByEmail(r.Context(), req.Email)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "邮箱未注册")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	user.PasswordHash = hash
	if err := s.store.UpdateUser(r.Context(), user); err != nil {
		writeStoreErr(w, err)
		return
	}
	// 重置后吊销全部刷新令牌。
	_ = s.refresh.Revoke(r.Context(), user.ID)
	s.log.Printf("[audit] password reset user=%d ip=%s", user.ID, clientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// issueTokens 签发 access+refresh 并响应。
func (s *Server) issueTokens(w http.ResponseWriter, r *http.Request, user *store.User, status int) {
	access, err := s.jwt.SignAccess(user.ID, user.Nickname, user.Tag)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	refresh, err := s.refresh.IssueWithUID(r.Context(), user.ID)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, status, tokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(auth.AccessTokenTTL / time.Second),
		User:         user,
	})
}

// normalizeEmail 邮箱小写去空白。
func normalizeEmail(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r != ' ' && r != '\t' {
			out = append(out, r)
		}
	}
	return strings.ToLower(string(out))
}
