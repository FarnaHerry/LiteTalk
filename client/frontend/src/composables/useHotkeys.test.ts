import { describe, expect, it } from 'vitest'
import { HOTKEYS, isTypingTarget, matchesHotkey, resolveHotkey } from './useHotkeys'

function key(k: string, ctrl = false, shift = false, alt = false) {
  return { ctrlKey: ctrl, shiftKey: shift, altKey: alt, key: k }
}

describe('matchesHotkey', () => {
  const mute = HOTKEYS.find((h) => h.id === 'mute')!
  const dual = HOTKEYS.find((h) => h.id === 'dual')!

  it('命中 Ctrl+M', () => {
    expect(matchesHotkey(key('m', true), mute)).toBe(true)
    expect(matchesHotkey(key('M', true), mute)).toBe(true) // 大小写不敏感
  })

  it('修饰键不符不命中', () => {
    expect(matchesHotkey(key('m'), mute)).toBe(false) // 无 Ctrl
    expect(matchesHotkey(key('m', true, true), mute)).toBe(false) // 多 Shift
    expect(matchesHotkey(key('d', true), mute)).toBe(false) // 键不符
  })

  it('双修饰键 Ctrl+Shift+M 不与 Ctrl+M 混淆', () => {
    expect(matchesHotkey(key('m', true, true), dual)).toBe(true)
    expect(matchesHotkey(key('m', true), dual)).toBe(false)
    expect(matchesHotkey(key('m', true, true), mute)).toBe(false)
  })
})

describe('isTypingTarget', () => {
  it('输入类元素不触发', () => {
    expect(isTypingTarget({ tagName: 'INPUT' })).toBe(true)
    expect(isTypingTarget({ tagName: 'TEXTAREA' })).toBe(true)
    expect(isTypingTarget({ tagName: 'SELECT' })).toBe(true)
    expect(isTypingTarget({ tagName: 'DIV', isContentEditable: true })).toBe(true)
  })

  it('普通元素触发', () => {
    expect(isTypingTarget({ tagName: 'DIV' })).toBe(false)
    expect(isTypingTarget({ tagName: 'BUTTON' })).toBe(false)
    expect(isTypingTarget(null)).toBe(false)
    expect(isTypingTarget(undefined)).toBe(false)
    expect(isTypingTarget({})).toBe(false)
  })
})

describe('resolveHotkey', () => {
  it('解析全部注册组合', () => {
    expect(resolveHotkey(key('m', true))).toBe('mute')
    expect(resolveHotkey(key('d', true))).toBe('deafen')
    expect(resolveHotkey(key('m', true, true))).toBe('dual')
    expect(resolveHotkey(key('k', true))).toBe('search')
    expect(resolveHotkey(key(',', true))).toBe('settings')
  })

  it('未注册组合返回空串', () => {
    expect(resolveHotkey(key('x', true))).toBe('')
    expect(resolveHotkey(key('m'))).toBe('')
  })
})
