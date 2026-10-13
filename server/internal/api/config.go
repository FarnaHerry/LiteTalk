package api

import "net/http"

// handleConfig GET /api/v1/config（免鉴权）：
// 返回服务端能力开关，客户端据此适配注册/找回密码 UI。
// mail_configured 表示发信通道是否为真实投递（非 Console 打印）。
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"auth_mode":       s.authMode,
		"mail_configured": s.mailConfigured(),
	})
}
