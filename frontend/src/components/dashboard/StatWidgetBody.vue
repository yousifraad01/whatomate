<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { TrendingUp, TrendingDown, Minus } from 'lucide-vue-next'
import { Skeleton } from '@/components/ui/skeleton'
import Sparkline from '@/components/dashboard/Sparkline.vue'
import { formatWidgetNumber, formatChange, trendOf } from '@/lib/dashboard'

const props = withDefaults(defineProps<{
  value: number
  change?: number
  prevValue?: number
  showChange?: boolean
  /** e.g. "from last month" */
  comparisonLabel: string
  /** Daily values for the sparkline (omitted or short → no sparkline) */
  series?: number[]
  /** Tailwind text colour class that tints the sparkline */
  accentClass?: string
  /** Phone tier: single line, no sparkline */
  compact?: boolean
  loading?: boolean
  unavailable?: boolean
}>(), {
  change: 0,
  prevValue: undefined,
  showChange: true,
  series: () => [],
  accentClass: 'text-primary',
  compact: false,
  loading: false,
  unavailable: false,
})

const { t } = useI18n()

const trend = computed(() => trendOf(props.change))

const trendIcon = computed(() => (trend.value === 'up' ? TrendingUp : trend.value === 'down' ? TrendingDown : Minus))

const trendPillClass = computed(() => {
  switch (trend.value) {
    case 'up':
      return 'bg-success text-success-foreground'
    case 'down':
      return 'bg-danger text-danger-foreground'
    default:
      return 'bg-muted text-muted-foreground'
  }
})

const trendSrText = computed(() =>
  trend.value === 'up' ? t('dashboard.increase') : trend.value === 'down' ? t('dashboard.decrease') : t('dashboard.noChange'),
)

const prevTitle = computed(() =>
  props.prevValue === undefined ? undefined : t('dashboard.comparedTo', { label: formatWidgetNumber(props.prevValue) }),
)

const hasSeries = computed(() => !props.compact && props.series.length > 1)
</script>

<template>
  <div v-if="loading" aria-busy="true" :class="compact ? 'space-y-1.5' : 'space-y-2'">
    <Skeleton :class="compact ? 'h-7 w-16' : 'h-9 w-24'" />
    <Skeleton class="h-4 w-28" />
  </div>

  <div v-else-if="unavailable" class="flex h-full flex-col justify-center gap-1">
    <span :class="['font-semibold leading-none tracking-tight text-muted-foreground', compact ? 'text-2xl' : 'text-3xl']" aria-hidden="true">—</span>
    <span class="text-xs text-muted-foreground">{{ t('dashboard.dataUnavailable') }}</span>
  </div>

  <!-- Phone tier: value and trend on one line -->
  <div v-else-if="compact" class="flex h-full items-end justify-between gap-2">
    <Transition name="counter-fade" mode="out-in">
      <span :key="value" class="text-2xl font-semibold leading-none tracking-tight tabular-nums text-foreground">
        {{ formatWidgetNumber(value) }}
      </span>
    </Transition>
    <span
      v-if="showChange"
      :class="['inline-flex shrink-0 items-center gap-0.5 rounded-full px-1.5 py-0.5 text-xs font-medium tabular-nums', trendPillClass]"
      :title="prevTitle"
    >
      <component :is="trendIcon" class="h-3 w-3" aria-hidden="true" />
      <span class="sr-only">{{ trendSrText }}</span>
      {{ formatChange(change) }}
      <span class="sr-only">{{ comparisonLabel }}</span>
    </span>
  </div>

  <!-- Tablet and desktop tiers -->
  <div v-else class="flex h-full flex-col">
    <Transition name="counter-fade" mode="out-in">
      <div :key="value" class="text-3xl font-semibold leading-none tracking-tight tabular-nums text-foreground">
        {{ formatWidgetNumber(value) }}
      </div>
    </Transition>

    <div v-if="showChange" class="mt-2 flex flex-wrap items-center gap-x-1.5 gap-y-1 text-xs text-muted-foreground">
      <span
        :class="['inline-flex items-center gap-0.5 rounded-full px-1.5 py-0.5 font-medium tabular-nums', trendPillClass]"
        :title="prevTitle"
      >
        <component :is="trendIcon" class="h-3 w-3" aria-hidden="true" />
        <span class="sr-only">{{ trendSrText }}</span>
        {{ formatChange(change) }}
      </span>
      <span>{{ comparisonLabel }}</span>
    </div>

    <!-- The sparkline takes whatever height is left in the card -->
    <div v-if="hasSeries" :class="['mt-auto min-h-[14px] pt-2', accentClass]">
      <Sparkline :points="series" :height="30" />
    </div>
  </div>
</template>
