<script setup lang="ts">
import { computed } from 'vue'

/**
 * Minimal inline sparkline. Purely decorative: the value and trend are
 * conveyed as text next to it, so the SVG is hidden from assistive tech.
 * Colour comes from `currentColor`, so wrap it in an element with the
 * accent text colour you want.
 */
const props = withDefaults(defineProps<{
  points: number[]
  /** Rendered height in px; the width follows the container. */
  height?: number
}>(), {
  height: 32,
})

const W = 100
const H = 32
const PAD = 2

const coords = computed(() => {
  const pts = props.points
  if (pts.length < 2) return []
  let min = Math.min(...pts)
  let max = Math.max(...pts)
  if (max === min) {
    // Flat series: an all-zero series sits on the baseline, any other
    // constant value runs through the vertical middle.
    if (max === 0) {
      max = 1
    } else {
      min -= 1
      max += 1
    }
  }
  const stepX = W / (pts.length - 1)
  return pts.map((v, i) => ({
    x: i * stepX,
    y: H - PAD - ((v - min) / (max - min)) * (H - PAD * 2),
  }))
})

const linePath = computed(() =>
  coords.value.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(2)} ${p.y.toFixed(2)}`).join(' ')
)

const areaPath = computed(() => {
  const c = coords.value
  if (c.length === 0) return ''
  return `${linePath.value} L${W} ${H} L0 ${H} Z`
})
</script>

<template>
  <svg
    v-if="coords.length > 0"
    class="block w-full"
    :style="{ height: `${height}px` }"
    :viewBox="`0 0 ${W} ${H}`"
    preserveAspectRatio="none"
    aria-hidden="true"
    focusable="false"
  >
    <path :d="areaPath" fill="currentColor" fill-opacity="0.12" stroke="none" />
    <path
      :d="linePath"
      fill="none"
      stroke="currentColor"
      stroke-width="1.75"
      stroke-linejoin="round"
      stroke-linecap="round"
      vector-effect="non-scaling-stroke"
    />
  </svg>
</template>
