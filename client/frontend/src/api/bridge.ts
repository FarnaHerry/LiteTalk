// bridge 封装 Wails Go 绑定（window.go.main.App.*）。
// 浏览器独立开发（vite dev 无 Go 后端）时走 mock 桩，行为可联调 UI。
import type { AuthResult, PeerView, UserProfile } from './types'

type GoApp = {
  GetVersion(): Promise<string>
  GetHibernateOnMinimize(): Promise<boolean>
  SetHibernateOnMinimize(enabled: boolean): Promise<void>
  WindowMinimize(): Promise<void>
  WindowMaximize(): Promise<void>
  WindowClose(): Promise<void>
  SendEmailCode(email: string, purpose: string): Promise<void>
  Login(email: string, password: string): Promise<AuthResult>
  Register(nickname: string, email: string, password: string, code: string): Promise<AuthResult>
  ResetPassword(email: string, code: string, newPassword: string): Promise<void>
  Logout(): Promise<void>
  GetMe(): Promise<UserProfile>
  ListAudioDevices(): Promise<{ capture: AudioDeviceDTO[]; playback: AudioDeviceDTO[] }>
  SetAudioDevices(captureID: string, playbackID: string): Promise<void>
  JoinChannel(channelID: string): Promise<void>
  LeaveChannel(): Promise<void>
  SetMute(muted: boolean): Promise<void>
  SetDeafen(deafened: boolean): Promise<void>
  SetPeerVolume(userID: string, gain: number): Promise<void>
  SendChat(channelID: string, content: string): Promise<void>
  GetVoiceMembers(): Promise<PeerView[]>
}

export interface AudioDeviceDTO {
  id: string
  name: string
  is_default: boolean
}

declare global {
  interface Window {
    go?: { main?: { App?: GoApp } }
    runtime?: {
      EventsOn(name: string, cb: (...args: unknown[]) => void): void
      EventsOff(name: string): void
    }
  }
}

// mock：浏览器开发环境的桩（登录一律成功、设备返回虚拟项）。
const mockApp: GoApp = {
  async GetVersion() { return 'dev' },
  async GetHibernateOnMinimize() { return false },
  async SetHibernateOnMinimize() {},
  async WindowMinimize() {},
  async WindowMaximize() {},
  async WindowClose() {},
  async SendEmailCode() {},
  async Login(email: string) {
    return {
      access_token: 'dev-token', refresh_token: 'dev-refresh',
      user: { id: email.split('@')[0] || 'dev', nickname: email.split('@')[0] || 'dev', tag: '0001', email },
    }
  },
  async Register(_n: string, email: string) {
    return mockApp.Login(email, "mock")
  },
  async ResetPassword() {},
  async Logout() {},
  async GetMe() {
    return { id: 'dev', nickname: 'dev', tag: '0001', email: 'dev@example.com' }
  },
  async SetAudioDevices() {},
  async ListAudioDevices() {
    return {
      capture: [{ id: 'default', name: '默认麦克风', is_default: true }],
      playback: [{ id: 'default', name: '默认扬声器', is_default: true }],
    }
  },
  async JoinChannel() {},
  async LeaveChannel() {},
  async SetMute() {},
  async SetDeafen() {},
  async SetPeerVolume() {},
  async SendChat() {},
  async GetVoiceMembers() { return [] },
}

export function goApp(): GoApp {
  return window.go?.main?.App ?? mockApp
}

export function isNative(): boolean {
  return !!window.go?.main?.App
}

// eventsOn Wails 事件订阅（浏览器环境静默）。
export function eventsOn(name: string, cb: (...args: unknown[]) => void) {
  window.runtime?.EventsOn(name, cb)
}

// eventsOff 取消事件订阅（休眠态清理用）。
export function eventsOff(name: string) {
  window.runtime?.EventsOff(name)
}
