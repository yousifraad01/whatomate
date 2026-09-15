<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { ArrowRight, GripVertical, Pencil, Trash2 } from 'lucide-vue-next'
import type { Component } from 'vue'

withDefaults(defineProps<{
  title: string
  description?: string
  icon?: Component
  /** Tailwind classes for the icon tile (background + text colour) */
  iconClass?: string
  canEdit?: boolean
  canDelete?: boolean
  dragMode?: boolean
  /** Whether the body gets the standard horizontal padding */
  padded?: boolean
  /** Tighter paddings for the phone tier */
  compact?: boolean
  /** Optional "view all" destination shown in the header */
  to?: string
  linkLabel?: string
}>(), {
  canEdit: false,
  canDelete: false,
  dragMode: false,
  padded: true,
  compact: false,
})

const emit = defineEmits<{
  edit: []
  delete: []
}>()

const { t } = useI18n()
</script>

<template>
  <!-- .card-depth is the shared widget surface (also targeted by e2e tests). -->
  <section
    :class="[
      'card-depth group relative flex h-full flex-col overflow-hidden rounded-lg border transition-colors',
      dragMode ? 'border-dashed border-primary/40' : 'hover:border-muted-foreground/40',
    ]"
    :aria-label="title"
  >
    <header :class="['flex items-start justify-between gap-3', compact ? 'px-4 pb-1 pt-3' : 'px-5 pb-2 pt-4']">
      <div class="min-w-0">
        <h2 :class="['font-medium text-muted-foreground', compact ? 'line-clamp-2 text-xs leading-4' : 'truncate text-sm']">{{ title }}</h2>
        <p v-if="description && !compact" class="truncate text-xs text-muted-foreground/80">{{ description }}</p>
      </div>
      <div class="flex shrink-0 items-center gap-1">
        <RouterLink
          v-if="to && !dragMode"
          :to="to"
          class="inline-flex h-7 items-center gap-1 rounded-md px-2 text-xs font-medium text-primary hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {{ linkLabel || t('dashboard.viewAll') }}
          <ArrowRight class="h-3.5 w-3.5" aria-hidden="true" />
        </RouterLink>
        <div
          v-if="dragMode"
          class="widget-drag-handle cursor-grab rounded p-1 text-muted-foreground hover:bg-accent active:cursor-grabbing"
          role="img"
          :aria-label="t('dashboard.dragToReorder')"
        >
          <GripVertical class="h-4 w-4" />
        </div>
        <template v-else-if="canEdit || canDelete">
          <!-- Revealed on hover and on keyboard focus so the controls are reachable without a mouse -->
          <Button
            v-if="canEdit"
            variant="ghost"
            size="icon"
            class="h-7 w-7 text-muted-foreground opacity-0 transition-opacity focus-visible:opacity-100 group-focus-within:opacity-100 group-hover:opacity-100"
            :title="t('dashboard.editWidgetTooltip')"
            :aria-label="t('dashboard.editWidgetTooltip')"
            @click.stop="emit('edit')"
          >
            <Pencil class="h-3.5 w-3.5" aria-hidden="true" />
          </Button>
          <Button
            v-if="canDelete"
            variant="ghost"
            size="icon"
            class="h-7 w-7 text-muted-foreground opacity-0 transition-opacity hover:text-destructive focus-visible:opacity-100 group-focus-within:opacity-100 group-hover:opacity-100"
            :title="t('dashboard.deleteWidgetTooltip')"
            :aria-label="t('dashboard.deleteWidgetTooltip')"
            @click.stop="emit('delete')"
          >
            <Trash2 class="h-3.5 w-3.5" aria-hidden="true" />
          </Button>
        </template>
        <div
          v-if="icon"
          :class="[
            'flex items-center justify-center rounded-md',
            compact ? 'h-6 w-6' : 'h-8 w-8',
            iconClass || 'bg-muted text-muted-foreground',
          ]"
          aria-hidden="true"
        >
          <component :is="icon" :class="compact ? 'h-3 w-3' : 'h-4 w-4'" />
        </div>
      </div>
    </header>
    <div :class="['min-h-0 flex-1', padded && (compact ? 'px-4 pb-3' : 'px-5 pb-4')]">
      <slot />
    </div>
  </section>
</template>
