<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { messageStatusMeta } from '@/lib/status'

const props = defineProps<{
  status: string | undefined
}>()

const { t } = useI18n()
const meta = computed(() => messageStatusMeta(props.status))
</script>

<template>
  <!-- The tick icon is decorative; the visually hidden text carries the
       delivery state for screen readers (WCAG 1.4.1 / 1.1.1). -->
  <span class="inline-flex items-center" :title="t(meta.labelKey)">
    <component
      :is="meta.icon"
      :class="['status-icon h-4 w-4', props.status === 'read' && 'status-read', props.status === 'failed' && 'text-destructive']"
      aria-hidden="true"
    />
    <span class="sr-only">{{ t(meta.labelKey) }}</span>
  </span>
</template>
