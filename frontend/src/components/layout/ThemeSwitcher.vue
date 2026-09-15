<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Sun, Moon, Monitor } from 'lucide-vue-next'
import { useColorMode, type ColorMode } from '@/composables/useColorMode'

const { colorMode, setColorMode } = useColorMode()
const { t } = useI18n()

const options: Array<{ value: ColorMode; icon: typeof Sun; labelKey: string }> = [
  { value: 'light', icon: Sun, labelKey: 'userMenu.themeLight' },
  { value: 'dark', icon: Moon, labelKey: 'userMenu.themeDark' },
  { value: 'system', icon: Monitor, labelKey: 'userMenu.themeSystem' },
]
</script>

<template>
  <div class="flex gap-0.5 px-1.5 py-1" role="radiogroup" :aria-label="t('userMenu.theme')">
    <Button
      v-for="option in options"
      :key="option.value"
      variant="ghost"
      size="icon"
      class="h-7 w-7"
      :class="colorMode === option.value && 'bg-accent text-accent-foreground'"
      :aria-checked="colorMode === option.value"
      :aria-label="t(option.labelKey)"
      :title="t(option.labelKey)"
      role="radio"
      @click="setColorMode(option.value)"
    >
      <component :is="option.icon" class="h-3.5 w-3.5" aria-hidden="true" />
    </Button>
  </div>
</template>
