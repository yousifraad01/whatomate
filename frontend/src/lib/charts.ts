/**
 * Centralized Chart.js setup
 * Import this module in components that need charts to ensure registration happens once
 */
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'

// Register Chart.js components once
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

// Set default options for better tooltip behavior
// This makes tooltips show when hovering near data points, not just exactly on them
ChartJS.defaults.interaction.mode = 'index'
ChartJS.defaults.interaction.intersect = false
ChartJS.defaults.font.family = "Inter, ui-sans-serif, system-ui, sans-serif"

/**
 * Series colours in reading order, resolved from the design tokens so charts
 * follow the active theme. Falls back to the dark palette outside the DOM.
 */
export function chartPalette(count = 6): string[] {
  const style = typeof window !== 'undefined' ? getComputedStyle(document.documentElement) : null
  const fallback = ['#22c55e', '#38bdf8', '#fbbf24', '#a78bfa', '#f87171', '#22d3ee']
  const colors: string[] = []
  for (let i = 1; i <= Math.max(1, count); i++) {
    const token = style?.getPropertyValue(`--chart-${((i - 1) % 6) + 1}`).trim()
    colors.push(token ? `hsl(${token})` : fallback[(i - 1) % fallback.length])
  }
  return colors
}

/**
 * Apply theme-aware defaults for axis labels, grid lines and legends so text
 * keeps adequate contrast in both light and dark mode. Call whenever the
 * theme changes; existing chart instances re-read defaults on update.
 */
export function applyChartTheme(isDark: boolean) {
  ChartJS.defaults.color = isDark ? 'hsl(0 0% 72%)' : 'hsl(0 0% 32%)'
  ChartJS.defaults.borderColor = isDark ? 'hsl(0 0% 20%)' : 'hsl(0 0% 88%)'
}

// Re-export chart components for convenience
export { Line, Bar, Pie, Doughnut } from 'vue-chartjs'
export { ChartJS }
