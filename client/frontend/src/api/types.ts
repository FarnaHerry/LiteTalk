// 类型定义（与 Go 侧 state 包对齐）。
export interface UserProfile {
  id: string
  nickname: string
  tag: string
  email: string
}

export interface AuthResult {
  access_token: string
  refresh_token: string
  user: UserProfile
}

export interface PeerView {
  user_id: string
  nickname: string
  speaking: boolean
  mute: boolean
  deafen: boolean
}

export interface ChatMessage {
  id: string
  channel_id: string
  user_id: string
  nickname: string
  content: string
  ts: string
}
