<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Input } from '@/components/ui/input'
import { Search, X } from 'lucide-vue-next'

const model = defineModel<string>({ default: '' })

const props = defineProps<{
  placeholder?: string
  /** Accessible name; defaults to the placeholder text */
  label?: string
  class?: string
}>()

const { t } = useI18n()

const effectivePlaceholder = computed(() => props.placeholder || `${t('common.search')}…`)
const accessibleName = computed(() => props.label || props.placeholder || t('common.search'))

function clear() {
  model.value = ''
}
</script>

<template>
  <div class="relative" :class="props.class" role="search">
    <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
    <Input
      v-model="model"
      type="search"
      :placeholder="effectivePlaceholder"
      :aria-label="accessibleName"
      class="pl-9 pr-8 [&::-webkit-search-cancel-button]:hidden"
    />
    <button
      v-if="model"
      type="button"
      class="absolute right-2 top-1/2 flex h-5 w-5 -translate-y-1/2 items-center justify-center rounded-full text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      :aria-label="t('common.clearSearch')"
      @click="clear"
    >
      <X class="h-3 w-3" aria-hidden="true" />
    </button>
  </div>
</template>
