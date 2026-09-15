<script setup lang="ts">
import { ref, onMounted, computed, watch, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import { GridLayout, GridItem } from 'grid-layout-plus'
import { useMediaQuery } from '@vueuse/core'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { EmptyState } from '@/components/ui/empty-state'
import { widgetsService, type DashboardWidget, type WidgetData, type LayoutItem } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import {
  MessageSquare,
  Users,
  Bot,
  Send,
  RefreshCw,
  Clock,
  LayoutDashboard,
  Plus,
  BarChart3,
  FileText,
  X,
  GripVertical,
  Megaphone,
  Settings,
  Contact,
  Workflow,
  Key,
  UserX,
  MessageSquareText,
  Webhook,
  ShieldCheck,
  Zap,
  Shield,
  LineChart,
  Tags
} from 'lucide-vue-next'
// Centralized Chart.js setup (registered once)
import { Line, Bar, Pie, chartPalette, applyChartTheme } from '@/lib/charts'
import { PageHeader, DateRangePicker, DeleteConfirmDialog } from '@/components/shared'
import WidgetCard from '@/components/dashboard/WidgetCard.vue'
import StatWidgetBody from '@/components/dashboard/StatWidgetBody.vue'
import { useDateRange } from '@/composables/useDateRange'
import { useAppToast } from '@/composables/useAppToast'
import { useColorMode } from '@/composables/useColorMode'
import { currentLocale } from '@/lib/utils'
import { directionMeta, messageStatusMeta } from '@/lib/status'
import { deriveLayout, widgetKind, formatWidgetNumber, LAYOUT_COLS, type GridLayoutItem } from '@/lib/dashboard'

const { success, error: showError } = useAppToast()
const { t } = useI18n()
const authStore = useAuthStore()
const { isDark } = useColorMode()

// Chart text/grid colours follow the theme
watch(isDark, (dark) => applyChartTheme(dark), { immediate: true })

// Permission checks
const canCreateWidget = computed(() => authStore.hasPermission('analytics', 'write'))
const canEditWidget = computed(() => authStore.hasPermission('analytics', 'write'))
const canDeleteWidget = computed(() => authStore.hasPermission('analytics', 'delete'))

// Widgets state
const widgets = ref<DashboardWidget[]>([])
const widgetData = ref<Record<string, WidgetData>>({})
// Widgets whose query failed server-side; rendered as "unavailable", never as 0
const widgetErrors = ref<Record<string, string>>({})

const isLoading = ref(true)
const isWidgetDataLoading = ref(false)
// When the widget data was last fetched successfully (shown in the header)
const lastUpdatedAt = ref<Date | null>(null)

// Widget builder state
const isWidgetDialogOpen = ref(false)
const isEditMode = ref(false)
const editingWidgetId = ref<string | null>(null)
const isSavingWidget = ref(false)

// Delete dialog state
const deleteDialogOpen = ref(false)
const widgetToDelete = ref<DashboardWidget | null>(null)
const isDeletingWidget = ref(false)

const dataSources = ref<Array<{ name: string; label: string; fields: string[] }>>([])
const metrics = ref<string[]>([])
const displayTypes = ref<string[]>([])
const operators = ref<Array<{ value: string; label: string }>>([])

const widgetForm = ref({
  name: '',
  description: '',
  data_source: '',
  metric: 'count',
  field: '',
  filters: [] as Array<{ field: string; operator: string; value: string }>,
  display_type: 'number',
  chart_type: '',
  group_by_field: '',
  show_change: true,
  color: 'blue',
  size: 'small',
  config: {} as Record<string, any>,
  is_shared: false
})

// Selected shortcuts for shortcuts widget creation
const selectedShortcuts = ref<string[]>([])

// Shortcut registry
const SHORTCUT_REGISTRY = computed(() => ({
  chat: { label: t('dashboard.startChat'), to: '/chat', icon: MessageSquare },
  campaigns: { label: t('nav.campaigns'), to: '/campaigns', icon: Megaphone },
  templates: { label: t('nav.templates'), to: '/templates', icon: FileText },
  chatbot: { label: t('nav.chatbot'), to: '/chatbot', icon: Bot },
  contacts: { label: t('nav.contacts'), to: '/settings/contacts', icon: Contact },
  flows: { label: t('nav.flows'), to: '/flows', icon: Workflow },
  transfers: { label: t('nav.transfers'), to: '/chatbot/transfers', icon: UserX },
  agentAnalytics: { label: t('nav.agentAnalytics'), to: '/analytics/agents', icon: BarChart3 },
  metaInsights: { label: t('nav.metaInsights'), to: '/analytics/meta-insights', icon: LineChart },
  settings: { label: t('nav.settings'), to: '/settings', icon: Settings },
  accounts: { label: t('nav.accounts'), to: '/settings/accounts', icon: Users },
  cannedResponses: { label: t('nav.cannedResponses'), to: '/settings/canned-responses', icon: MessageSquareText },
  tags: { label: t('nav.tags'), to: '/settings/tags', icon: Tags },
  teams: { label: t('nav.teams'), to: '/settings/teams', icon: Users },
  users: { label: t('nav.users'), to: '/settings/users', icon: Users },
  roles: { label: t('nav.roles'), to: '/settings/roles', icon: Shield },
  apiKeys: { label: t('nav.apiKeys'), to: '/settings/api-keys', icon: Key },
  webhooks: { label: t('nav.webhooks'), to: '/settings/webhooks', icon: Webhook },
  customActions: { label: t('nav.customActions'), to: '/settings/custom-actions', icon: Zap },
  sso: { label: t('nav.sso'), to: '/settings/sso', icon: ShieldCheck },
}))

interface ShortcutEntry {
  label: string
  to: string
  icon: Component
}

function shortcutFor(key: string): ShortcutEntry | undefined {
  return (SHORTCUT_REGISTRY.value as Record<string, ShortcutEntry>)[key]
}

// Color options: a muted tint for the widget icon tile (icons are decorative,
// so the tint only needs 3:1 against the card, which these all clear)
const colorOptions = computed(() => [
  { value: 'blue', label: t('dashboard.colorBlue'), bg: 'bg-sky-500/15', text: 'text-sky-600 light:text-sky-700', swatch: 'bg-sky-500' },
  { value: 'green', label: t('dashboard.colorGreen'), bg: 'bg-emerald-500/15', text: 'text-emerald-600 light:text-emerald-700', swatch: 'bg-emerald-500' },
  { value: 'purple', label: t('dashboard.colorPurple'), bg: 'bg-violet-500/15', text: 'text-violet-500 light:text-violet-700', swatch: 'bg-violet-500' },
  { value: 'orange', label: t('dashboard.colorOrange'), bg: 'bg-amber-500/15', text: 'text-amber-600 light:text-amber-700', swatch: 'bg-amber-500' },
  { value: 'red', label: t('dashboard.colorRed'), bg: 'bg-red-500/15', text: 'text-red-500 light:text-red-700', swatch: 'bg-red-500' },
  { value: 'cyan', label: t('dashboard.colorCyan'), bg: 'bg-cyan-500/15', text: 'text-cyan-600 light:text-cyan-700', swatch: 'bg-cyan-500' }
])

// Chart type options
const chartTypeOptions = computed(() => [
  { value: 'line', label: t('dashboard.chartLine') },
  { value: 'bar', label: t('dashboard.chartBar') },
  { value: 'pie', label: t('dashboard.chartPie') }
])

// Series colours come from the design tokens and follow the active theme
const chartColors = computed(() => {
  void isDark.value
  return chartPalette(8)
})

function withAlpha(color: string, alpha: number): string {
  return color.startsWith('hsl(') ? color.replace(')', ` / ${alpha})`) : color
}

const getChartComponentData = (widget: DashboardWidget) => {
  const data = widgetData.value[widget.id]
  if (!data) return { labels: [], datasets: [] }

  const chartData = data.chart_data || []
  const dataPoints = data.data_points || []
  const groupedSeries = data.grouped_series
  const colors = chartColors.value

  // Grouped line chart: multiple datasets from grouped_series
  if (widget.chart_type === 'line' && groupedSeries && groupedSeries.datasets.length > 0) {
    return {
      labels: groupedSeries.labels,
      datasets: groupedSeries.datasets.map((ds, i) => ({
        label: ds.label,
        data: ds.data,
        borderColor: colors[i % colors.length],
        backgroundColor: withAlpha(colors[i % colors.length], 0.12),
        fill: false,
        tension: 0.3
      }))
    }
  }

  // Bar/Pie with group_by uses data_points (group → count)
  if (widget.chart_type === 'pie') {
    const source = dataPoints.length > 0 ? dataPoints : chartData
    return {
      labels: source.map((d: { label: string }) => d.label),
      datasets: [{
        data: source.map((d: { value: number }) => d.value),
        backgroundColor: colors.slice(0, source.length),
        borderWidth: 0
      }]
    }
  }

  if (widget.chart_type === 'bar' && dataPoints.length > 0) {
    return {
      labels: dataPoints.map((d: { label: string }) => d.label),
      datasets: [{
        label: widget.name,
        data: dataPoints.map((d: { value: number }) => d.value),
        backgroundColor: dataPoints.map((_: any, i: number) => colors[i % colors.length]),
        borderWidth: 0
      }]
    }
  }

  // Default: line and bar charts use time-series chart_data
  const colorIndex: Record<string, number> = { green: 0, blue: 1, orange: 2, purple: 3, red: 4, cyan: 5 }
  const borderColor = colors[colorIndex[widget.color] ?? 0]

  return {
    labels: chartData.map((d: { label: string }) => d.label),
    datasets: [{
      label: widget.name,
      data: chartData.map((d: { value: number }) => d.value),
      borderColor,
      backgroundColor: widget.chart_type === 'bar' ? borderColor : withAlpha(borderColor, 0.12),
      fill: widget.chart_type === 'line',
      tension: 0.3
    }]
  }
}

const lineBarChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: true, position: 'top' as const }
  },
  scales: {
    y: { beginAtZero: true }
  }
}

const pieChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'bottom' as const }
  }
}

// Time range filter
const {
  selectedRange,
  customDateRange,
  isDatePickerOpen,
  dateRange,
  formatDateRangeDisplay,
  applyCustomRange: applyCustomRangeBase,
} = useDateRange({ storageKey: 'dashboard' })

const comparisonPeriodLabel = computed(() => {
  switch (selectedRange.value) {
    case 'today':
      return t('dashboard.fromYesterday')
    case '7days':
      return t('dashboard.fromPrevious7Days')
    case '30days':
      return t('dashboard.fromPrevious30Days')
    case 'this_month':
      return t('dashboard.fromLastMonth')
    case 'custom':
      return t('dashboard.fromPreviousPeriod')
    default:
      return t('dashboard.fromPreviousPeriod')
  }
})

// The reporting period is stated explicitly (with its time basis) rather
// than implied by the preset name.
const periodLabel = computed(() => {
  const { from, to } = dateRange.value
  const fmt = new Intl.DateTimeFormat(currentLocale(), { day: 'numeric', month: 'short', year: 'numeric' })
  const start = fmt.format(new Date(from + 'T00:00:00'))
  const end = fmt.format(new Date(to + 'T00:00:00'))
  const range = from === to ? start : `${start} – ${end}`
  const base = `${t('dashboard.reportingPeriod')}: ${range} · ${t('common.localTime')}`
  if (!lastUpdatedAt.value) return base
  const time = new Intl.DateTimeFormat(currentLocale(), { hour: 'numeric', minute: '2-digit' }).format(lastUpdatedAt.value)
  return `${base} · ${t('dashboard.lastUpdated', { time })}`
})

const formatNumber = formatWidgetNumber

// Daily values behind a stat card sparkline
const seriesFor = (id: string): number[] => (widgetData.value[id]?.chart_data || []).map(p => p.value)

// Destination of a list widget "view all" link
const viewAllTarget = (widget: DashboardWidget): string | undefined => {
  switch (widget.data_source) {
    case 'messages':
      return '/chat'
    case 'contacts':
      return '/settings/contacts'
    case 'campaigns':
      return '/campaigns'
    case 'transfers':
      return '/chatbot/transfers'
    default:
      return undefined
  }
}

const formatTime = (dateStr: string): string => {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return t('dashboard.justNow')
  if (diffMins < 60) return t('dashboard.minutesAgo', { count: diffMins })
  if (diffHours < 24) return t('dashboard.hoursAgo', { count: diffHours })
  return t('dashboard.daysAgo', { count: diffDays })
}

const getWidgetColor = (color: string) => {
  return colorOptions.value.find(c => c.value === color) || colorOptions.value[0]
}

const getWidgetIcon = (dataSource: string) => {
  switch (dataSource) {
    case 'messages':
      return MessageSquare
    case 'contacts':
      return Users
    case 'sessions':
      return Bot
    case 'campaigns':
      return Send
    case 'transfers':
      return Users
    default:
      return BarChart3
  }
}

// Whether the widget has any data to draw (as opposed to an empty result)
const hasChartData = (id: string) => {
  const d = widgetData.value[id]
  return !!d && ((d.chart_data?.length || 0) > 0 || (d.data_points?.length || 0) > 0 || (d.grouped_series?.datasets?.length || 0) > 0)
}

const isUnavailable = (id: string) => !!widgetErrors.value[id] || (!isWidgetDataLoading.value && !widgetData.value[id])

// Grid layout state
const GRID_COLS = 12
const GRID_ROW_HEIGHT = 40
const GRID_MARGIN: [number, number] = [16, 16]

const isDragMode = ref(false)
const gridLayout = ref<Array<{ i: string; x: number; y: number; w: number; h: number }>>([])

// Layout tiers. The saved 12-column layout is only used from the xl
// breakpoint; tablets and small laptops get a 6-column and phones a 2-column layout derived
// from it in reading order. Derived layouts are never persisted and cannot
// be edited, so drag mode is a desktop-only feature.
const isNarrow = useMediaQuery('(max-width: 767px)')
const isTablet = useMediaQuery('(min-width: 768px) and (max-width: 1279px)')
const isDesktop = computed(() => !isNarrow.value && !isTablet.value)
const displayCols = computed(() =>
  isNarrow.value ? LAYOUT_COLS.phone : isTablet.value ? LAYOUT_COLS.tablet : LAYOUT_COLS.desktop,
)
const gridMargin = computed<[number, number]>(() => (isNarrow.value ? [12, 12] : GRID_MARGIN))
const kindOfWidget = (id: string) => widgetKind(getWidgetById(id)?.display_type || 'number')
// On desktop the saved layout is passed through by reference: a fresh array
// would make the grid re-emit layout-updated on every render and loop.
const displayLayout = computed<GridLayoutItem[]>(() =>
  isDesktop.value ? gridLayout.value : deriveLayout(gridLayout.value, displayCols.value, kindOfWidget),
)

// Shortcut tiles: three per row when the widget is wide enough
const shortcutColumns = (item: GridLayoutItem) =>
  !isNarrow.value && item.w / displayCols.value >= 0.66 ? 'grid-cols-3' : 'grid-cols-2'

const isChartWidget = (widget: DashboardWidget) => widget.display_type === 'chart'
const isTableWidget = (widget: DashboardWidget) => widget.display_type === 'table'
const isShortcutsWidget = (widget: DashboardWidget) => widget.display_type === 'shortcuts'
const isNumberWidget = (widget: DashboardWidget) => !isChartWidget(widget) && !isTableWidget(widget) && !isShortcutsWidget(widget)

const getWidgetById = (id: string): DashboardWidget | undefined => {
  return widgets.value.find(w => w.id === id)
}

const computeGridLayout = (widgetList: DashboardWidget[]) => {
  const layout: Array<{ i: string; x: number; y: number; w: number; h: number }> = []

  // Separate positioned (grid_w > 0) from legacy (grid_w === 0) widgets
  const positioned = widgetList.filter(w => w.grid_w > 0)
  const legacy = widgetList.filter(w => w.grid_w === 0)

  // Add positioned widgets as-is
  for (const w of positioned) {
    layout.push({ i: w.id, x: w.grid_x, y: w.grid_y, w: w.grid_w, h: w.grid_h })
  }

  // Auto-position legacy widgets
  if (legacy.length > 0) {
    // Find the max y used by positioned widgets to place legacy below
    let nextY = 0
    if (positioned.length > 0) {
      nextY = Math.max(...positioned.map(w => w.grid_y + w.grid_h))
    }

    let curX = 0
    let curY = nextY

    // Number widgets first, then chart/table/shortcuts widgets
    const legacyNumber = legacy.filter(w => !['chart', 'table', 'shortcuts'].includes(w.display_type))
    const legacyLarge = legacy.filter(w => ['chart', 'table', 'shortcuts'].includes(w.display_type))

    for (const w of legacyNumber) {
      const itemW = 3
      const itemH = 3
      if (curX + itemW > GRID_COLS) {
        curX = 0
        curY += itemH
      }
      layout.push({ i: w.id, x: curX, y: curY, w: itemW, h: itemH })
      curX += itemW
    }

    // Move to next row for large widgets
    if (legacyNumber.length > 0 && legacyLarge.length > 0) {
      curX = 0
      curY += 3
    }

    for (const w of legacyLarge) {
      let itemW = 6
      let itemH = 5
      if (w.display_type === 'table' || w.display_type === 'shortcuts') {
        itemW = 6
        itemH = 8
      }
      if (curX + itemW > GRID_COLS) {
        curX = 0
        curY += itemH
      }
      layout.push({ i: w.id, x: curX, y: curY, w: itemW, h: itemH })
      curX += itemW
    }
  }

  return layout
}

// Rebuild grid layout when widgets change
watch(widgets, (val) => {
  gridLayout.value = computeGridLayout(val)
}, { immediate: true })

// Debounced layout save
let layoutSaveTimer: ReturnType<typeof setTimeout> | null = null

const persistLayout = async () => {
  const layoutItems: LayoutItem[] = gridLayout.value.map(item => ({
    id: item.i,
    grid_x: item.x,
    grid_y: item.y,
    grid_w: item.w,
    grid_h: item.h
  }))
  try {
    await widgetsService.saveLayout(layoutItems)
  } catch (error: any) {
    showError(t('common.error'), error.response?.data?.message || t('dashboard.saveLayoutFailed'))
  }
}

const onLayoutUpdate = (newLayout: Array<{ i: string; x: number; y: number; w: number; h: number }>) => {
  if (!isDesktop.value) return
  // Ignore no-op updates so a watcher cycle cannot form
  if (JSON.stringify(newLayout) === JSON.stringify(gridLayout.value)) return
  gridLayout.value = newLayout
  if (!isDragMode.value) return
  if (layoutSaveTimer) clearTimeout(layoutSaveTimer)
  layoutSaveTimer = setTimeout(persistLayout, 500)
}

// Save immediately when exiting drag mode
watch(isDragMode, (newVal, oldVal) => {
  if (oldVal && !newVal) {
    // Toggled off — save now
    if (layoutSaveTimer) {
      clearTimeout(layoutSaveTimer)
      layoutSaveTimer = null
    }
    persistLayout()
  }
})

watch(isDesktop, (desktop) => {
  if (!desktop) isDragMode.value = false
})

const availableFields = computed(() => {
  if (!widgetForm.value.data_source) return []
  const source = dataSources.value.find(s => s.name === widgetForm.value.data_source)
  return source?.fields || []
})

// Fetch data
const fetchWidgets = async () => {
  try {
    const response = await widgetsService.list()
    widgets.value = (response.data as any).data?.widgets || []
  } catch (error) {
    console.error('Failed to load widgets:', error)
    widgets.value = []
  }
}

const fetchWidgetData = async () => {
  if (widgets.value.length === 0) return

  isWidgetDataLoading.value = true
  try {
    const { from, to } = dateRange.value
    const response = await widgetsService.getAllData({ from, to })
    const payload = (response.data as any).data || {}
    widgetData.value = payload.data || {}
    widgetErrors.value = payload.errors || {}
    lastUpdatedAt.value = new Date()
  } catch (error) {
    console.error('Failed to load widget data:', error)
    widgetData.value = {}
    // Every widget is unavailable, not zero
    widgetErrors.value = Object.fromEntries(widgets.value.map(w => [w.id, 'request_failed']))
  } finally {
    isWidgetDataLoading.value = false
  }
}

const fetchDataSources = async () => {
  try {
    const response = await widgetsService.getDataSources()
    const data = (response.data as any).data || response.data
    dataSources.value = data.data_sources || []
    metrics.value = data.metrics || []
    displayTypes.value = data.display_types || []
    operators.value = data.operators || []
  } catch (error) {
    console.error('Failed to load data sources:', error)
  }
}

const fetchDashboardData = async () => {
  isLoading.value = true
  try {
    await Promise.all([
      fetchWidgets(),
      fetchDataSources()
    ])
    await fetchWidgetData()
  } finally {
    isLoading.value = false
  }
}

const applyCustomRange = () => {
  applyCustomRangeBase()
  fetchWidgetData()
}

// Widget CRUD
const openAddWidgetDialog = () => {
  isEditMode.value = false
  editingWidgetId.value = null
  widgetForm.value = {
    name: '',
    description: '',
    data_source: '',
    metric: 'count',
    field: '',
    filters: [],
    display_type: 'number',
    chart_type: '',
    group_by_field: '',
    show_change: true,
    color: 'blue',
    size: 'small',
    config: {},
    is_shared: false
  }
  selectedShortcuts.value = []
  isWidgetDialogOpen.value = true
}

const openEditWidgetDialog = (widget: DashboardWidget) => {
  isEditMode.value = true
  editingWidgetId.value = widget.id
  widgetForm.value = {
    name: widget.name,
    description: widget.description,
    data_source: widget.data_source,
    metric: widget.metric,
    field: widget.field,
    filters: [...widget.filters],
    display_type: widget.display_type,
    chart_type: widget.chart_type,
    group_by_field: widget.group_by_field || '',
    show_change: widget.show_change,
    color: widget.color || 'blue',
    size: widget.size,
    config: widget.config || {},
    is_shared: widget.is_shared
  }
  // Populate selectedShortcuts from config
  if (widget.display_type === 'shortcuts' && widget.config?.shortcuts) {
    selectedShortcuts.value = [...widget.config.shortcuts as string[]]
  } else {
    selectedShortcuts.value = []
  }
  isWidgetDialogOpen.value = true
}

const addFilter = () => {
  widgetForm.value.filters.push({ field: '', operator: 'equals', value: '' })
}

const removeFilter = (index: number) => {
  widgetForm.value.filters.splice(index, 1)
}

const saveWidget = async () => {
  if (isSavingWidget.value) return
  const isShortcuts = widgetForm.value.display_type === 'shortcuts'

  if (!widgetForm.value.name) {
    showError(t('dashboard.validationError'), t('dashboard.nameRequired'))
    return
  }

  if (!isShortcuts && !widgetForm.value.data_source) {
    showError(t('dashboard.validationError'), t('dashboard.dataSourceRequired'))
    return
  }

  // Clean up empty filters
  const cleanFilters = widgetForm.value.filters.filter(f => f.field && f.operator && f.value)

  // Build config
  let config: Record<string, any> = { ...widgetForm.value.config }
  if (isShortcuts) {
    config = { shortcuts: [...selectedShortcuts.value] }
  }

  const payload = {
    name: widgetForm.value.name,
    description: widgetForm.value.description,
    data_source: widgetForm.value.data_source,
    metric: widgetForm.value.metric,
    field: widgetForm.value.field,
    filters: cleanFilters,
    display_type: widgetForm.value.display_type,
    chart_type: widgetForm.value.chart_type,
    group_by_field: widgetForm.value.group_by_field,
    show_change: widgetForm.value.show_change,
    color: widgetForm.value.color,
    size: widgetForm.value.size,
    config,
    is_shared: widgetForm.value.is_shared
  }

  isSavingWidget.value = true
  try {
    if (isEditMode.value && editingWidgetId.value) {
      await widgetsService.update(editingWidgetId.value, payload)
      success(t('common.updatedSuccess', { resource: t('resources.Widget') }))
    } else {
      await widgetsService.create(payload)
      success(t('common.createdSuccess', { resource: t('resources.Widget') }))
    }
    isWidgetDialogOpen.value = false
    await fetchWidgets()
    await fetchWidgetData()
  } catch (error: any) {
    showError(t('common.error'), error.response?.data?.message || t('common.failedSave', { resource: t('resources.widget') }))
  } finally {
    isSavingWidget.value = false
  }
}

const openDeleteDialog = (widget: DashboardWidget) => {
  widgetToDelete.value = widget
  deleteDialogOpen.value = true
}

const confirmDeleteWidget = async () => {
  if (!widgetToDelete.value || isDeletingWidget.value) return

  isDeletingWidget.value = true
  try {
    await widgetsService.delete(widgetToDelete.value.id)
    success(t('common.deletedSuccess', { resource: t('resources.Widget') }))
    deleteDialogOpen.value = false
    widgetToDelete.value = null
    await fetchWidgets()
    await fetchWidgetData()
  } catch (error: any) {
    showError(t('common.error'), error.response?.data?.message || t('common.failedDelete', { resource: t('resources.widget') }))
  } finally {
    isDeletingWidget.value = false
  }
}

// Watch for range changes
watch(selectedRange, (newValue) => {
  if (newValue !== 'custom') {
    fetchWidgetData()
  }
})

// Set default chart_type when display_type changes to chart
watch(() => widgetForm.value.display_type, (newVal) => {
  if (newVal === 'chart' && !widgetForm.value.chart_type) {
    widgetForm.value.chart_type = 'line'
  }
  if (newVal !== 'chart') {
    widgetForm.value.chart_type = ''
  }
  if (newVal !== 'chart' && newVal !== 'table') {
    widgetForm.value.group_by_field = ''
  }
  if (newVal === 'shortcuts') {
    widgetForm.value.data_source = ''
    widgetForm.value.metric = 'count'
  }
})

onMounted(() => {
  fetchDashboardData()
})
</script>

<template>
  <div class="flex h-full flex-col bg-background">
    <PageHeader :title="$t('dashboard.title')" :description="periodLabel" :icon="LayoutDashboard">
      <template #actions>
        <Button v-if="canCreateWidget" variant="outline" size="sm" @click="openAddWidgetDialog">
          <Plus class="h-4 w-4" aria-hidden="true" />
          {{ $t('dashboard.addWidget') }}
        </Button>

        <Button
          v-if="canEditWidget && widgets.length > 1 && isDesktop"
          :variant="isDragMode ? 'active' : 'outline'"
          size="sm"
          :aria-pressed="isDragMode"
          @click="isDragMode = !isDragMode"
        >
          <GripVertical class="h-4 w-4" aria-hidden="true" />
          {{ isDragMode ? $t('common.done') : $t('dashboard.editLayout') }}
        </Button>

        <Button
          variant="outline"
          size="icon"
          class="h-9 w-9"
          :aria-label="$t('dashboard.refresh')"
          :title="$t('dashboard.refresh')"
          :disabled="isWidgetDataLoading || widgets.length === 0"
          @click="fetchWidgetData"
        >
          <RefreshCw :class="['h-4 w-4', isWidgetDataLoading && 'animate-spin']" aria-hidden="true" />
        </Button>

        <DateRangePicker
          v-model:selected-range="selectedRange"
          v-model:custom-date-range="customDateRange"
          v-model:is-date-picker-open="isDatePickerOpen"
          :format-date-range-display="formatDateRangeDisplay"
          @apply-custom="applyCustomRange"
        />
      </template>
    </PageHeader>

    <!-- Content -->
    <ScrollArea class="flex-1">
      <div class="space-y-6 p-4 sm:p-6">
        <!-- Loading Skeleton -->
        <div v-if="isLoading" class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4" role="status" :aria-label="$t('common.loading')">
          <div v-for="i in 4" :key="i" class="rounded-lg border border-border bg-card p-5">
            <div class="flex items-center justify-between pb-2">
              <Skeleton class="h-4 w-24" />
              <Skeleton class="h-8 w-8 rounded-md" />
            </div>
            <div class="pt-2">
              <Skeleton class="mb-2 h-8 w-20" />
              <Skeleton class="h-3 w-32" />
            </div>
          </div>
        </div>

        <!-- Empty dashboard -->
        <EmptyState
          v-else-if="widgets.length === 0"
          :icon="LayoutDashboard"
          :title="$t('dashboard.noWidgets')"
          :description="$t('dashboard.noWidgetsDesc')"
          class="rounded-lg border border-dashed border-border"
        >
          <template v-if="canCreateWidget" #action>
            <Button size="sm" @click="openAddWidgetDialog">
              <Plus class="h-4 w-4" aria-hidden="true" />
              {{ $t('dashboard.addWidget') }}
            </Button>
          </template>
        </EmptyState>

        <!-- Widget Grid Layout -->
        <GridLayout
          v-else-if="gridLayout.length > 0"
          :layout="displayLayout"
          :col-num="displayCols"
          :row-height="GRID_ROW_HEIGHT"
          :margin="gridMargin"
          :is-draggable="isDragMode && isDesktop"
          :is-resizable="isDragMode && isDesktop"
          :vertical-compact="true"
          :use-css-transforms="true"
          @layout-updated="onLayoutUpdate"
        >
          <GridItem
            v-for="item in displayLayout"
            :key="item.i"
            :i="item.i"
            :x="item.x"
            :y="item.y"
            :w="item.w"
            :h="item.h"
            :min-w="isDesktop ? 2 : 1"
            :min-h="2"
            drag-allow-from=".widget-drag-handle"
          >
            <template v-if="getWidgetById(item.i)">
              <!-- Number widget -->
              <WidgetCard
                v-if="isNumberWidget(getWidgetById(item.i)!)"
                :title="getWidgetById(item.i)!.name"
                :icon="getWidgetIcon(getWidgetById(item.i)!.data_source)"
                :icon-class="`${getWidgetColor(getWidgetById(item.i)!.color).bg} ${getWidgetColor(getWidgetById(item.i)!.color).text}`"
                :can-edit="canEditWidget"
                :can-delete="canDeleteWidget"
                :drag-mode="isDragMode"
                :compact="isNarrow"
                @edit="openEditWidgetDialog(getWidgetById(item.i)!)"
                @delete="openDeleteDialog(getWidgetById(item.i)!)"
              >
                <StatWidgetBody
                  :value="widgetData[item.i]?.value ?? 0"
                  :change="widgetData[item.i]?.change ?? 0"
                  :prev-value="widgetData[item.i]?.prev_value"
                  :show-change="getWidgetById(item.i)!.show_change"
                  :comparison-label="comparisonPeriodLabel"
                  :series="seriesFor(item.i)"
                  :accent-class="getWidgetColor(getWidgetById(item.i)!.color).text"
                  :compact="isNarrow"
                  :loading="isWidgetDataLoading"
                  :unavailable="isUnavailable(item.i)"
                />
              </WidgetCard>

              <!-- Chart widget -->
              <WidgetCard
                v-else-if="isChartWidget(getWidgetById(item.i)!)"
                :title="getWidgetById(item.i)!.name"
                :description="getWidgetById(item.i)!.description"
                :icon="getWidgetIcon(getWidgetById(item.i)!.data_source)"
                :icon-class="`${getWidgetColor(getWidgetById(item.i)!.color).bg} ${getWidgetColor(getWidgetById(item.i)!.color).text}`"
                :can-edit="canEditWidget"
                :can-delete="canDeleteWidget"
                :drag-mode="isDragMode"
                :compact="isNarrow"
                @edit="openEditWidgetDialog(getWidgetById(item.i)!)"
                @delete="openDeleteDialog(getWidgetById(item.i)!)"
              >
                <div class="h-full min-h-0">
                  <Skeleton v-if="isWidgetDataLoading" class="h-full w-full" />
                  <div v-else-if="isUnavailable(item.i)" class="flex h-full items-center justify-center text-sm text-muted-foreground">
                    {{ $t('dashboard.dataUnavailable') }}
                  </div>
                  <template v-else-if="hasChartData(item.i)">
                    <Line v-if="getWidgetById(item.i)!.chart_type === 'line'" :data="getChartComponentData(getWidgetById(item.i)!)" :options="lineBarChartOptions" :aria-label="getWidgetById(item.i)!.name" role="img" />
                    <Bar v-else-if="getWidgetById(item.i)!.chart_type === 'bar'" :data="getChartComponentData(getWidgetById(item.i)!)" :options="lineBarChartOptions" :aria-label="getWidgetById(item.i)!.name" role="img" />
                    <Pie v-else-if="getWidgetById(item.i)!.chart_type === 'pie'" :data="getChartComponentData(getWidgetById(item.i)!)" :options="pieChartOptions" :aria-label="getWidgetById(item.i)!.name" role="img" />
                  </template>
                  <div v-else class="flex h-full items-center justify-center text-sm text-muted-foreground">
                    {{ $t('common.noData') }}
                  </div>
                </div>
              </WidgetCard>

              <!-- Table widget -->
              <WidgetCard
                v-else-if="isTableWidget(getWidgetById(item.i)!)"
                :title="getWidgetById(item.i)!.name"
                :description="getWidgetById(item.i)!.description"
                :can-edit="canEditWidget"
                :can-delete="canDeleteWidget"
                :drag-mode="isDragMode"
                :compact="isNarrow"
                :to="viewAllTarget(getWidgetById(item.i)!)"
                :padded="false"
                @edit="openEditWidgetDialog(getWidgetById(item.i)!)"
                @delete="openDeleteDialog(getWidgetById(item.i)!)"
              >
                <div :class="['h-full min-h-0 overflow-auto', isNarrow ? 'px-4 pb-3' : 'px-5 pb-4']">
                  <Skeleton v-if="isWidgetDataLoading" class="h-full w-full" />
                  <div v-else-if="isUnavailable(item.i)" class="flex h-full items-center justify-center text-sm text-muted-foreground">
                    {{ $t('dashboard.dataUnavailable') }}
                  </div>
                  <!-- Grouped table (group_by set) -->
                  <table v-else-if="getWidgetById(item.i)!.group_by_field && widgetData[item.i]?.data_points?.length" class="w-full text-sm">
                    <caption class="sr-only">{{ getWidgetById(item.i)!.name }}</caption>
                    <thead>
                      <tr class="border-b border-border">
                        <th scope="col" class="py-2 text-left text-xs font-medium uppercase tracking-wide text-muted-foreground">{{ getWidgetById(item.i)!.group_by_field }}</th>
                        <th scope="col" class="py-2 text-right text-xs font-medium uppercase tracking-wide text-muted-foreground">{{ $t('dashboard.count') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="dp in widgetData[item.i]?.data_points" :key="dp.label" class="border-b border-border/60">
                        <td class="py-2 text-foreground">{{ dp.label }}</td>
                        <td class="py-2 text-right font-medium tabular-nums text-foreground">{{ formatNumber(dp.value) }}</td>
                      </tr>
                    </tbody>
                  </table>
                  <!-- Row list (no group_by) -->
                  <ul v-else-if="widgetData[item.i]?.table_rows?.length" class="divide-y divide-border">
                    <li
                      v-for="row in widgetData[item.i]?.table_rows"
                      :key="row.id"
                      class="flex items-start gap-3 py-2.5"
                    >
                      <div
                        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-xs font-medium text-muted-foreground"
                        aria-hidden="true"
                      >
                        {{ row.label.split(' ').map((n: string) => n[0]).join('').slice(0, 2).toUpperCase() }}
                      </div>
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center justify-between gap-2">
                          <p class="truncate text-sm font-medium text-foreground">{{ row.label }}</p>
                          <span class="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
                            <Clock class="h-3 w-3" aria-hidden="true" />
                            {{ formatTime(row.created_at) }}
                          </span>
                        </div>
                        <p class="truncate text-sm text-muted-foreground">{{ row.sub_label }}</p>
                        <div class="mt-1 flex items-center gap-1.5">
                          <Badge v-if="row.direction" :variant="directionMeta(row.direction).variant" class="px-1.5 py-0 text-[10px]">
                            {{ $t(directionMeta(row.direction).labelKey) }}
                          </Badge>
                          <Badge v-if="row.status" :variant="messageStatusMeta(row.status).variant" class="px-1.5 py-0 text-[10px]">
                            {{ $t(messageStatusMeta(row.status).labelKey) }}
                          </Badge>
                        </div>
                      </div>
                    </li>
                  </ul>
                  <div v-else class="flex h-full items-center justify-center text-sm text-muted-foreground">
                    {{ $t('common.noData') }}
                  </div>
                </div>
              </WidgetCard>

              <!-- Shortcuts widget -->
              <WidgetCard
                v-else-if="isShortcutsWidget(getWidgetById(item.i)!)"
                :title="getWidgetById(item.i)!.name"
                :description="getWidgetById(item.i)!.description"
                :can-edit="canEditWidget"
                :can-delete="canDeleteWidget"
                :drag-mode="isDragMode"
                :compact="isNarrow"
                :padded="false"
                @edit="openEditWidgetDialog(getWidgetById(item.i)!)"
                @delete="openDeleteDialog(getWidgetById(item.i)!)"
              >
                <nav :class="['h-full min-h-0 overflow-y-auto', isNarrow ? 'px-4 pb-3' : 'px-5 pb-4']" :aria-label="getWidgetById(item.i)!.name">
                  <div :class="['grid gap-2 pt-1', shortcutColumns(item)]">
                    <template v-for="key in (getWidgetById(item.i)!.config?.shortcuts || [])" :key="key">
                      <RouterLink
                        v-if="shortcutFor(key as string)"
                        :to="shortcutFor(key as string)!.to"
                        class="card-interactive flex items-center gap-3 rounded-md border border-border bg-background px-3 py-2.5 text-sm font-medium text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                      >
                        <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground" aria-hidden="true">
                          <component :is="shortcutFor(key as string)!.icon" class="h-4 w-4" />
                        </span>
                        <span class="truncate">{{ shortcutFor(key as string)!.label }}</span>
                      </RouterLink>
                    </template>
                  </div>
                </nav>
              </WidgetCard>
            </template>
          </GridItem>
        </GridLayout>
      </div>
    </ScrollArea>

    <!-- Widget Dialog -->
    <Dialog v-model:open="isWidgetDialogOpen">
      <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle>{{ isEditMode ? $t('dashboard.editWidget') : $t('dashboard.createWidget') }}</DialogTitle>
          <DialogDescription>
            {{ $t('dashboard.widgetDialogDesc') }}
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-4 py-2" @submit.prevent="saveWidget">
          <!-- Name -->
          <div class="space-y-2">
            <Label for="widget-name">{{ $t('dashboard.widgetName') }} <span class="text-destructive" aria-hidden="true">*</span></Label>
            <Input id="widget-name" v-model="widgetForm.name" :placeholder="$t('dashboard.widgetNamePlaceholder')" required />
          </div>

          <!-- Description -->
          <div class="space-y-2">
            <Label for="widget-description">{{ $t('dashboard.widgetDescription') }}</Label>
            <Textarea id="widget-description" v-model="widgetForm.description" :placeholder="$t('dashboard.widgetDescriptionPlaceholder')" :rows="2" />
          </div>

          <!-- Data Source (hidden for shortcuts) -->
          <div v-if="widgetForm.display_type !== 'shortcuts'" class="space-y-2">
            <Label for="widget-source">{{ $t('dashboard.dataSource') }} <span class="text-destructive" aria-hidden="true">*</span></Label>
            <Select :model-value="widgetForm.data_source" @update:model-value="(val) => widgetForm.data_source = String(val)">
              <SelectTrigger id="widget-source">
                <SelectValue :placeholder="$t('dashboard.selectDataSource')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="source in dataSources" :key="source.name" :value="source.name">
                  {{ source.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Metric (hidden for shortcuts and table) -->
          <div v-if="widgetForm.display_type !== 'shortcuts' && widgetForm.display_type !== 'table'" class="space-y-2">
            <Label for="widget-metric">{{ $t('dashboard.metric') }}</Label>
            <Select :model-value="widgetForm.metric" @update:model-value="(val) => widgetForm.metric = String(val)">
              <SelectTrigger id="widget-metric">
                <SelectValue :placeholder="$t('dashboard.selectMetric')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="count">{{ $t('dashboard.metricCount') }}</SelectItem>
                <SelectItem value="sum">{{ $t('dashboard.metricSum') }}</SelectItem>
                <SelectItem value="avg">{{ $t('dashboard.metricAverage') }}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Display Type -->
          <div class="space-y-2">
            <Label for="widget-display">{{ $t('dashboard.displayType') }}</Label>
            <Select :model-value="widgetForm.display_type" @update:model-value="(val) => widgetForm.display_type = String(val)">
              <SelectTrigger id="widget-display">
                <SelectValue :placeholder="$t('dashboard.selectDisplayType')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="number">{{ $t('dashboard.displayNumber') }}</SelectItem>
                <SelectItem value="chart">{{ $t('dashboard.displayChart') }}</SelectItem>
                <SelectItem value="table">{{ $t('dashboard.displayTable') }}</SelectItem>
                <SelectItem value="shortcuts">{{ $t('dashboard.displayShortcuts') }}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Chart Type (visible when display type is chart) -->
          <div v-if="widgetForm.display_type === 'chart'" class="space-y-2">
            <Label for="widget-chart-type">{{ $t('dashboard.chartType') }}</Label>
            <Select :model-value="widgetForm.chart_type" @update:model-value="(val) => widgetForm.chart_type = String(val)">
              <SelectTrigger id="widget-chart-type">
                <SelectValue :placeholder="$t('dashboard.selectChartType')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="ct in chartTypeOptions" :key="ct.value" :value="ct.value">
                  {{ ct.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Group By (visible when display type is chart or table, and data source is selected) -->
          <div v-if="(widgetForm.display_type === 'chart' || widgetForm.display_type === 'table') && widgetForm.data_source" class="space-y-2">
            <Label for="widget-group-by">{{ $t('dashboard.groupBy') }}</Label>
            <Select :model-value="widgetForm.group_by_field || 'none'" @update:model-value="(val) => widgetForm.group_by_field = val === 'none' ? '' : String(val)">
              <SelectTrigger id="widget-group-by">
                <SelectValue :placeholder="$t('dashboard.noneTimeSeries')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">{{ $t('dashboard.noneTimeSeries') }}</SelectItem>
                <SelectItem v-for="field in availableFields" :key="field" :value="field">
                  {{ field }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Shortcuts selector (only for shortcuts display type) -->
          <fieldset v-if="widgetForm.display_type === 'shortcuts'" class="space-y-2">
            <legend class="text-sm font-medium">{{ $t('dashboard.selectShortcuts') }}</legend>
            <div class="max-h-64 space-y-1 overflow-y-auto pr-1">
              <label
                v-for="(shortcut, key) in SHORTCUT_REGISTRY"
                :key="key"
                class="flex cursor-pointer items-center gap-3 rounded-md p-2 hover:bg-accent"
              >
                <input
                  type="checkbox"
                  :value="key"
                  v-model="selectedShortcuts"
                  class="h-4 w-4 rounded border-input text-primary focus-visible:ring-2 focus-visible:ring-ring"
                />
                <span class="flex h-7 w-7 items-center justify-center rounded-md bg-muted text-muted-foreground" aria-hidden="true">
                  <component :is="shortcut.icon" class="h-4 w-4" />
                </span>
                <span class="text-sm text-foreground">{{ shortcut.label }}</span>
              </label>
            </div>
          </fieldset>

          <!-- Filters (hidden for shortcuts) -->
          <div v-if="widgetForm.display_type !== 'shortcuts'" class="space-y-2">
            <div class="flex items-center justify-between">
              <Label>{{ $t('dashboard.filters') }} ({{ widgetForm.filters.length }})</Label>
              <Button type="button" variant="outline" size="sm" @click.stop.prevent="addFilter">
                <Plus class="h-4 w-4" aria-hidden="true" />
                {{ $t('dashboard.addFilter') }}
              </Button>
            </div>
            <p v-if="!widgetForm.data_source && widgetForm.filters.length === 0" class="text-xs text-muted-foreground">
              {{ $t('dashboard.selectDataSourceFirst') }}
            </p>
            <div v-for="(filter, index) in widgetForm.filters" :key="index" class="flex items-center gap-2">
              <div class="flex-1">
                <Select :model-value="filter.field" @update:model-value="(val) => filter.field = String(val)">
                  <SelectTrigger class="w-full text-sm" :aria-label="$t('dashboard.field')">
                    <SelectValue :placeholder="$t('dashboard.field')" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="field in availableFields" :key="field" :value="field">
                      {{ field }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div class="w-36">
                <Select :model-value="filter.operator" @update:model-value="(val) => filter.operator = String(val)">
                  <SelectTrigger class="w-full text-sm" :aria-label="$t('dashboard.operator')">
                    <SelectValue :placeholder="$t('dashboard.operator')" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="op in operators" :key="op.value" :value="op.value">
                      {{ op.label }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Input v-model="filter.value" :placeholder="$t('dashboard.value')" :aria-label="$t('dashboard.value')" class="flex-1 text-sm" />
              <Button type="button" variant="ghost" size="icon" class="shrink-0 text-muted-foreground hover:text-destructive" :aria-label="$t('common.remove')" @click="removeFilter(index)">
                <X class="h-4 w-4" aria-hidden="true" />
              </Button>
            </div>
          </div>

          <!-- Color (hidden for shortcuts and table) -->
          <div v-if="widgetForm.display_type !== 'shortcuts' && widgetForm.display_type !== 'table'" class="space-y-2">
            <Label for="widget-color">{{ $t('dashboard.color') }}</Label>
            <Select :model-value="widgetForm.color" @update:model-value="(val) => widgetForm.color = String(val)">
              <SelectTrigger id="widget-color">
                <SelectValue :placeholder="$t('dashboard.selectColor')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="color in colorOptions" :key="color.value" :value="color.value">
                  <span class="flex items-center gap-2">
                    <span :class="['inline-block h-3 w-3 rounded-full', color.swatch]" aria-hidden="true"></span>
                    {{ color.label }}
                  </span>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Options -->
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div v-if="widgetForm.display_type === 'number' || widgetForm.display_type === 'percentage'" class="flex items-center gap-2">
              <Switch id="widget-show-change" v-model:checked="widgetForm.show_change" />
              <Label for="widget-show-change" class="font-normal">{{ $t('dashboard.showPercentChange') }}</Label>
            </div>
            <div class="flex items-center gap-2">
              <Switch id="widget-shared" v-model:checked="widgetForm.is_shared" />
              <Label for="widget-shared" class="font-normal">{{ $t('dashboard.shareWithTeam') }}</Label>
            </div>
          </div>
          <button type="submit" class="hidden" tabindex="-1" aria-hidden="true" />
        </form>

        <DialogFooter>
          <Button variant="outline" :disabled="isSavingWidget" @click="isWidgetDialogOpen = false">
            {{ $t('common.cancel') }}
          </Button>
          <Button :loading="isSavingWidget" @click="saveWidget">
            {{ isEditMode ? $t('common.update') : $t('common.create') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirmation Dialog -->
    <DeleteConfirmDialog
      v-model:open="deleteDialogOpen"
      :title="$t('dashboard.deleteWidgetTitle')"
      :description="$t('dashboard.deleteWidgetConfirm', { name: widgetToDelete?.name })"
      :is-submitting="isDeletingWidget"
      @confirm="confirmDeleteWidget"
    />
  </div>
</template>

<style>
/* Grid layout placeholder styling */
.vue-grid-item.vue-grid-placeholder {
  background: hsl(var(--primary) / 0.08) !important;
  border: 2px dashed hsl(var(--primary) / 0.5) !important;
  border-radius: 0.5rem;
}

/* Grid resize handle styling */
.vue-grid-item > .vue-resizable-handle {
  width: 20px;
  height: 20px;
  bottom: 4px;
  right: 4px;
  background: none;
  cursor: se-resize;
}

.vue-grid-item > .vue-resizable-handle::after {
  content: '';
  position: absolute;
  right: 4px;
  bottom: 4px;
  width: 8px;
  height: 8px;
  border-right: 2px solid hsl(var(--muted-foreground) / 0.6);
  border-bottom: 2px solid hsl(var(--muted-foreground) / 0.6);
  border-radius: 0 0 2px 0;
}

/* Ensure grid items don't overflow */
.vue-grid-item {
  transition: all 200ms ease;
}

/* Animated counter transition */
.counter-fade-enter-active,
.counter-fade-leave-active {
  transition: opacity 0.2s ease;
}
.counter-fade-enter-from,
.counter-fade-leave-to {
  opacity: 0;
}
</style>
