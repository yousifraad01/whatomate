import { describe, it, expect } from 'vitest'
import { deriveLayout, formatChange, trendOf, widgetKind, LAYOUT_COLS, type GridLayoutItem, type WidgetKind } from './dashboard'

const kinds: Record<string, WidgetKind> = {
  a: 'number',
  b: 'number',
  c: 'number',
  d: 'number',
  chart: 'chart',
  table: 'table',
}
const kindOf = (id: string) => kinds[id] ?? 'number'

// Four stat cards on the first row, a chart and a table below (desktop layout).
const base: GridLayoutItem[] = [
  { i: 'chart', x: 0, y: 3, w: 6, h: 5 },
  { i: 'table', x: 6, y: 3, w: 6, h: 8 },
  { i: 'b', x: 3, y: 0, w: 3, h: 3 },
  { i: 'a', x: 0, y: 0, w: 3, h: 3 },
  { i: 'd', x: 9, y: 0, w: 3, h: 3 },
  { i: 'c', x: 6, y: 0, w: 3, h: 3 },
]

describe('deriveLayout', () => {
  it('returns the saved layout in reading order on desktop', () => {
    const out = deriveLayout(base, LAYOUT_COLS.desktop, kindOf)
    expect(out.map(i => i.i)).toEqual(['a', 'b', 'c', 'd', 'chart', 'table'])
    expect(out[0]).toEqual({ i: 'a', x: 0, y: 0, w: 3, h: 3 })
  })

  it('packs two compact stat cards per row on phones and stacks the rest', () => {
    const out = deriveLayout(base, LAYOUT_COLS.phone, kindOf)
    expect(out).toEqual([
      { i: 'a', x: 0, y: 0, w: 1, h: 2 },
      { i: 'b', x: 1, y: 0, w: 1, h: 2 },
      { i: 'c', x: 0, y: 2, w: 1, h: 2 },
      { i: 'd', x: 1, y: 2, w: 1, h: 2 },
      { i: 'chart', x: 0, y: 4, w: 2, h: 5 },
      { i: 'table', x: 0, y: 9, w: 2, h: 8 },
    ])
  })

  it('uses half-width stat cards and full-width widgets on tablets', () => {
    const out = deriveLayout(base, LAYOUT_COLS.tablet, kindOf)
    expect(out.slice(0, 2)).toEqual([
      { i: 'a', x: 0, y: 0, w: 3, h: 3 },
      { i: 'b', x: 3, y: 0, w: 3, h: 3 },
    ])
    expect(out[4]).toEqual({ i: 'chart', x: 0, y: 6, w: 6, h: 5 })
    expect(out[5]).toEqual({ i: 'table', x: 0, y: 11, w: 6, h: 8 })
  })

  it('never overlaps items within a row', () => {
    for (const cols of [LAYOUT_COLS.phone, LAYOUT_COLS.tablet]) {
      const out = deriveLayout(base, cols, kindOf)
      for (const item of out) expect(item.x + item.w).toBeLessThanOrEqual(cols)
    }
  })

  it('does not mutate the input', () => {
    const copy = JSON.parse(JSON.stringify(base))
    deriveLayout(base, LAYOUT_COLS.phone, kindOf)
    expect(base).toEqual(copy)
  })
})

describe('formatting helpers', () => {
  it('formats change with sign and one decimal', () => {
    expect(formatChange(12.345)).toBe('+12.3%')
    expect(formatChange(-3)).toBe('−3.0%')
    expect(formatChange(0)).toBe('0.0%')
  })

  it('classifies trends', () => {
    expect(trendOf(5)).toBe('up')
    expect(trendOf(-0.1)).toBe('down')
    expect(trendOf(0)).toBe('flat')
    expect(trendOf(undefined)).toBe('flat')
  })

  it('maps display types to widget kinds', () => {
    expect(widgetKind('chart')).toBe('chart')
    expect(widgetKind('table')).toBe('table')
    expect(widgetKind('shortcuts')).toBe('shortcuts')
    expect(widgetKind('number')).toBe('number')
    expect(widgetKind('percentage')).toBe('number')
  })
})
