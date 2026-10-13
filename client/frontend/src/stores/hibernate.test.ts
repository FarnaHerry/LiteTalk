import { describe, expect, it } from 'vitest'
import { hibernateReduce } from './hibernate'

describe('hibernateReduce 状态机', () => {
  it('hibernate 事件进入休眠', () => {
    expect(hibernateReduce(false, 'hibernate')).toBe(true)
  })

  it('wake 事件唤醒', () => {
    expect(hibernateReduce(true, 'wake')).toBe(false)
  })

  it('toggle-setting 不改变状态', () => {
    expect(hibernateReduce(true, 'toggle-setting')).toBe(true)
    expect(hibernateReduce(false, 'toggle-setting')).toBe(false)
  })

  it('重复事件幂等', () => {
    expect(hibernateReduce(true, 'hibernate')).toBe(true)
    expect(hibernateReduce(false, 'wake')).toBe(false)
  })
})
