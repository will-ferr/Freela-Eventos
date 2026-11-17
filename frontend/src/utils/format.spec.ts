import { describe, expect, it } from 'vitest'
import { formatDateTime, formatGoDuration, formatLatency, shortCommit } from './format'

describe('formatLatency', () => {
  it.each([
    [0, '< 1 ms'],
    [12.4, '12 ms'],
    [999, '999 ms'],
    [1500, '1.5 s'],
  ])('%s ms → %s', (ms, expected) => {
    expect(formatLatency(ms)).toBe(expected)
  })
})

describe('formatGoDuration', () => {
  it.each([
    ['45s', '45s'],
    ['5m30s', '5min'],
    ['2h15m0s', '2h 15min'],
    ['50h0m0s', '2d 2h'],
    ['valor-estranho', 'valor-estranho'],
  ])('%s → %s', (input, expected) => {
    expect(formatGoDuration(input)).toBe(expected)
  })
})

describe('formatDateTime', () => {
  it('retorna traço para datas inválidas', () => {
    expect(formatDateTime('não é data')).toBe('—')
  })

  it('formata no padrão brasileiro', () => {
    expect(formatDateTime(new Date(2026, 10, 17, 9, 5, 7))).toBe('17/11/2026, 09:05:07')
  })
})

describe('shortCommit', () => {
  it('abrevia hashes e mantém outros valores', () => {
    expect(shortCommit('4bf92f3577b34da6a3ce929d0e0e4736')).toBe('4bf92f3')
    expect(shortCommit('none')).toBe('none')
  })
})
