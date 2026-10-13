// Package signal 实现轻音 LiteTalk 客户端的 WebSocket 信令层。
//
// 协议契约见 docs/落地方案.md §5.3：信封 {type, data, ts}，
// 服务端持有同一份小定义（双端各持一份，可接受重复）。
package signal

import (
	"encoding/json"
	"time"
)

// nowMillis 当前 Unix 毫秒（信封 ts）。
func nowMillis() int64 { return time.Now().UnixMilli() }

// 信令消息类型（与服务端逐字一致）。
const (
	TypeHello           = "hello"             // S→C 鉴权成功后下发全量初始数据
	TypeChannelJoin     = "channel.join"      // C→S 加入语音频道
	TypeChannelJoined   = "channel.joined"    // S→C 房间成员全量
	TypeChannelLeave    = "channel.leave"     // C→S 离开频道
	TypeChannelPeerLeft = "channel.peer_left" // S→C 有成员离开
	TypeRTCOffer        = "rtc.offer"         // 双向 SDP offer
	TypeRTCAnswer       = "rtc.answer"        // 双向 SDP answer
	TypeRTCIce          = "rtc.ice"           // 双向 trickle ICE
	TypeVoiceState      = "voice.state"       // C→S→广播 说话/静音/耳聋
	TypeChatSend        = "chat.send"         // C→S
	TypeChatMessage     = "chat.message"      // S→C
	TypePresenceUpdate  = "presence.update"   // S→C
	TypeFriendRequest   = "friend.request"    // S→C
	TypeFriendAccept    = "friend.accept"     // S→C
	TypeError           = "error"             // S→C
)

// Envelope 信令信封。Data 延迟解析，由订阅者按 Type 自行解码。
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
	TS   int64           `json:"ts"`
}

// NewEnvelope 打包一条待发消息（ts 取当前 Unix 毫秒）。
func NewEnvelope(msgType string, data any) (Envelope, error) {
	env := Envelope{Type: msgType, TS: nowMillis()}
	if data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			return Envelope{}, err
		}
		env.Data = raw
	}
	return env, nil
}

// Decode 解析 Data 到 v（Data 为空时返回 nil 错误，v 保持零值）。
func (e Envelope) Decode(v any) error {
	if len(e.Data) == 0 {
		return nil
	}
	return json.Unmarshal(e.Data, v)
}

// ---- 各消息 data 载荷 ----

// SDPData 是 rtc.offer / rtc.answer 的载荷。
type SDPData struct {
	SDP string `json:"sdp"`
}

// ICEData 是 rtc.ice 的载荷，字段与 WebRTC ICECandidateInit / pion 对齐。
type ICEData struct {
	Candidate        string  `json:"candidate"`
	SDPMid           string  `json:"sdpMid,omitempty"`
	SDPMLineIndex    *uint16 `json:"sdpMLineIndex,omitempty"`
	UsernameFragment string  `json:"usernameFragment,omitempty"`
}

// ChannelJoinData 是 channel.join / channel.leave 的载荷。
type ChannelJoinData struct {
	ChannelID string `json:"channel_id"`
}

// Member 语音房间成员（channel.joined 全量项）。
type Member struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Speaking bool   `json:"speaking"`
	Mute     bool   `json:"mute"`
	Deafen   bool   `json:"deafen"`
}

// ChannelJoinedData 是 channel.joined 的载荷（房间成员全量 + 自身）。
type ChannelJoinedData struct {
	ChannelID string   `json:"channel_id"`
	Members   []Member `json:"members"`
	Self      Member   `json:"self"`
}

// PeerLeftData 是 channel.peer_left 的载荷。
type PeerLeftData struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
}

// VoiceStateData 是 voice.state 的载荷。
// 上行（C→S）不带 UserID；下行广播（S→C）由服务端补上 UserID 标识来源。
// 三个状态字段均为可选增量：指针 nil 表示该字段本次未变。
type VoiceStateData struct {
	UserID   string `json:"user_id,omitempty"`
	Speaking *bool  `json:"speaking,omitempty"`
	Mute     *bool  `json:"mute,omitempty"`
	Deafen   *bool  `json:"deafen,omitempty"`
}

// ErrorData 是 error 消息的载荷。
// Code 为字符串枚举（bad_message / unauthorized / ...），与 server 对齐（审查 #3）。
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Bool 返回 b 的指针（构造可选增量字段用）。
func Bool(b bool) *bool { return &b }
