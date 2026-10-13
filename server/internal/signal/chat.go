// chat 文字聊天：chat.send（C→S）→ 校验/频控/权限 → 落库 → chat.message 广播（S→C）。
// wire 契约与客户端 client/internal/signal/protocol.go 对齐（ChatSendData / ChatMessageData）。
package signal

import (
	"context"
	"strconv"
	"time"

	"github.com/OpenALC/LiteTalk/server/internal/store"
)

// ChatStore chat 持久化最小接口（store.Store 满足；nil 时仅广播不落库）。
type ChatStore interface {
	CreateMessage(ctx context.Context, msg *store.Message) error
	ChannelByID(ctx context.Context, id int64) (*store.Channel, error)
	Member(ctx context.Context, serverID, userID int64) (*store.ServerMember, error)
}

// RateLimiter 发送频控接口（auth.RateLimit 满足；nil 时不限流）。
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// 聊天发送限制：2000 字、20 条/分（key 与 REST messages 共用 msg: 前缀，用户级统一配额）。
const (
	chatMaxRunes   = 2000
	chatRateLimit  = 20
	chatRateWindow = time.Minute
)

// handleChatSend 文字消息发送：内容校验 → 频控 → 权限（在房间 或 频道成员）→ 落库 → 广播。
// 广播目标为该频道房间内全体成员，并始终回执发送者（前端以广播替换本地回显）。
func (h *Hub) handleChatSend(c *client, d ChatSendData) {
	if d.ChannelID == "" {
		h.sendError(c, ErrCodeBadMessage, "chat.send 需要 channel_id")
		return
	}
	if n := len([]rune(d.Content)); n == 0 || n > chatMaxRunes {
		h.sendError(c, ErrCodeBadMessage, "消息长度需为 1–2000 字符")
		return
	}
	ctx := context.Background()
	if h.limits != nil {
		ok, err := h.limits.Allow(ctx, "msg:"+c.userID, chatRateLimit, chatRateWindow)
		if err != nil || !ok {
			h.sendError(c, ErrCodeRateLimited, "发送过于频繁")
			return
		}
	}

	// 权限二选一：当前在该房间（M2 dev 房间场景）；
	// 或 channel_id 为数字频道且本人是该频道所属服务器成员（文字频道场景）。
	inRoom := c.roomIDLocked() == d.ChannelID
	var msg *store.Message
	if chID, err := strconv.ParseInt(d.ChannelID, 10, 64); err == nil && h.chatStore != nil {
		if ch, err := h.chatStore.ChannelByID(ctx, chID); err == nil {
			if uid, err := strconv.ParseInt(c.userID, 10, 64); err == nil {
				if _, err := h.chatStore.Member(ctx, ch.ServerID, uid); err == nil {
					msg = &store.Message{ChannelID: chID, UserID: uid, Content: d.Content}
				}
			}
		}
	}
	if !inRoom && msg == nil {
		h.sendError(c, ErrCodeForbidden, "无权向该频道发送消息")
		return
	}

	out := ChatMessageData{
		ChannelID: d.ChannelID,
		Content:   d.Content,
		TS:        time.Now().UnixMilli(),
		User:      UserBrief{ID: c.userID, Nickname: c.username},
	}
	if msg != nil {
		if err := h.chatStore.CreateMessage(ctx, msg); err != nil {
			h.sendError(c, ErrCodeInternal, "消息保存失败")
			return
		}
		out.ID = strconv.FormatInt(msg.ID, 10)
		out.TS = msg.CreatedAt.UnixMilli()
	} else {
		// dev 房间（无 store 频道）：仅广播，id 用时间戳占位。
		out.ID = strconv.FormatInt(out.TS, 10)
	}

	// 广播：房间成员 + 发送者回执（不在房间时单独回执）。
	h.mu.Lock()
	var recips []*client
	if room := h.rooms[d.ChannelID]; room != nil {
		recips = room.snapshotLocked()
	}
	h.mu.Unlock()
	sawSelf := false
	for _, m := range recips {
		if m == c {
			sawSelf = true
		}
		h.sendJSON(m, TypeChatMessage, out)
	}
	if !sawSelf {
		h.sendJSON(c, TypeChatMessage, out)
	}
}
