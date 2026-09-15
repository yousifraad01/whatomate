<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertCircle } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

defineProps<{
  title?: string
  description?: string
  class?: HTMLAttributes['class']
  retryLabel?: string
}>()

defineEmits<{
  retry: []
}>()

const { t } = useI18n()
</script>

<template>
  <div
    :class="cn('flex flex-col items-center justify-center px-4 py-12 text-center', $props.class)"
    role="alert"
  >
    <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-danger text-danger-foreground" aria-hidden="true">
      <AlertCircle class="h-6 w-6" />
    </div>
    <h3 class="text-base font-semibold text-foreground">
      <slot name="title">{{ title || t('common.loadErrorTitle') }}</slot>
    </h3>
    <p class="mt-1 max-w-sm text-sm text-muted-foreground">
      <slot name="description">{{ description || t('common.loadErrorDescription') }}</slot>
    </p>
    <div class="mt-4">
      <slot name="action">
        <Button variant="outline" size="sm" @click="$emit('retry')">
          {{ retryLabel || t('common.retry') }}
        </Button>
      </slot>
    </div>
  </div>
</template>
