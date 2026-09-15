<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import type { Component } from "vue"
import { cn } from "@/lib/utils"

interface EmptyStateProps {
  class?: HTMLAttributes["class"]
  icon?: Component
  title?: string
  description?: string
}

const props = defineProps<EmptyStateProps>()
</script>

<template>
  <div
    :class="cn('flex flex-col items-center justify-center px-4 py-12 text-center', props.class)"
    role="status"
    aria-live="polite"
  >
    <div
      v-if="props.icon || $slots.icon"
      class="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground"
      aria-hidden="true"
    >
      <slot name="icon">
        <component :is="props.icon" class="h-6 w-6" />
      </slot>
    </div>
    <h3 v-if="props.title || $slots.title" class="text-base font-semibold text-foreground">
      <slot name="title">{{ props.title }}</slot>
    </h3>
    <p
      v-if="props.description || $slots.description"
      class="mt-1 max-w-sm text-sm text-muted-foreground"
    >
      <slot name="description">{{ props.description }}</slot>
    </p>
    <div v-if="$slots.action" class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>
