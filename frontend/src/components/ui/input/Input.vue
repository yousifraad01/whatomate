<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  modelValue?: string | number
  type?: string
  placeholder?: string
  disabled?: boolean
  class?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const modelValue = computed({
  get: () => props.modelValue?.toString() ?? '',
  set: (value) => emit('update:modelValue', value)
})
</script>

<template>
  <input
    v-model="modelValue"
    :type="type ?? 'text'"
    :placeholder="placeholder"
    :disabled="disabled"
    :class="cn(
      'flex h-9 w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm text-foreground shadow-sm transition-colors placeholder:text-muted-foreground hover:border-muted-foreground/50 focus-visible:outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring/40 disabled:cursor-not-allowed disabled:opacity-50',
      props.class
    )"
  />
</template>
