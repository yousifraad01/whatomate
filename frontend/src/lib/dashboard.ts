/**
 * Pure helpers for the dashboard: responsive layout derivation and
 * number/trend formatting. Kept free of Vue so they are unit-testable.
 */
import { formatNumber as formatLocaleNumber } from '@/lib/utils'

export interface GridLayoutItem {
  i: string
  x: number
  y: number
  w: number
  h: number
}

export type WidgetKind = 'number' | 'chart' | 'table' | 'shortcuts'

/** Column counts for each layout tier. The saved layout is always 12 columns. */
export const LAYOUT_COLS = {
  desktop: 12,
  tablet: 6,
  phone: 2,
} as const

export type LayoutTier = keyof typeof LAYOUT_COLS

/** Width and height (in grid units) a widget kind takes in a derived tier. */
function derivedSize(kind: WidgetKind, cols: number, base: GridLayoutItem): { w: number; h: number } {
  if (cols === LAYOUT_COLS.phone) {
    // Two compact stat cards per row; everything else spans the row.
    return kind === 'number' ? { w: 1, h: 2 } : { w: 2, h: base.h }
  }
  // Tablet: two stat cards per row, other widgets full width.
  return kind === 'number' ? { w: 3, h: base.h } : { w: cols, h: base.h }
}

/**
 * Derive a layout for a narrower column count from the saved 12-column one.
 * Widgets keep their desktop reading order (row by row, left to right) and
 * are packed into rows; the result is never persisted.
 */
export function deriveLayout(
  base: GridLayoutItem[],
  cols: number,
  kindOf: (id: string) => WidgetKind,
): GridLayoutItem[] {
  const ordered = [...base].sort((a, b) => a.y - b.y || a.x - b.x)
  if (cols >= LAYOUT_COLS.desktop) return ordered

  const out: GridLayoutItem[] = []
  let curX = 0
  let curY = 0
  let rowH = 0

  for (const item of ordered) {
    const { w, h } = derivedSize(kindOf(item.i), cols, item)
    if (curX + w > cols) {
      curX = 0
      curY += rowH
      rowH = 0
    }
    out.push({ i: item.i, x: curX, y: curY, w, h })
    curX += w
    rowH = Math.max(rowH, h)
  }
  return out
}

export function widgetKind(displayType: string): WidgetKind {
  switch (displayType) {
    case 'chart':
      return 'chart'
    case 'table':
      return 'table'
    case 'shortcuts':
      return 'shortcuts'
    default:
      return 'number'
  }
}

export type Trend = 'up' | 'down' | 'flat'

export function trendOf(change: number | undefined | null): Trend {
  const c = change ?? 0
  if (c > 0) return 'up'
  if (c < 0) return 'down'
  return 'flat'
}

/** Compact notation from 10k upwards, otherwise a plain integer. */
export function formatWidgetNumber(value: number): string {
  return formatLocaleNumber(
    value,
    Math.abs(value) >= 10000 ? { notation: 'compact', maximumFractionDigits: 1 } : { maximumFractionDigits: 0 },
  )
}

/** "+12.5%", "−3.0%" or "0.0%" with a real minus sign. */
export function formatChange(change: number): string {
  const sign = change > 0 ? '+' : change < 0 ? '−' : ''
  return `${sign}${Math.abs(change).toFixed(1)}%`
}
