<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowUp } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const props = withDefaults(defineProps<{
  target?: string
  threshold?: number
}>(), {
  threshold: 300,
})

const { t } = useI18n()
const isVisible = ref(false)
let scrollEl: HTMLElement | null = null

function onScroll() {
  if (scrollEl) {
    isVisible.value = scrollEl.scrollTop > props.threshold
  }
}

function scrollToTop() {
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  scrollEl?.scrollTo({ top: 0, behavior: reduceMotion ? 'auto' : 'smooth' })
}

onMounted(() => {
  scrollEl = props.target
    ? document.querySelector(props.target)
    : document.querySelector('main')
  scrollEl?.addEventListener('scroll', onScroll, { passive: true })
})

onUnmounted(() => {
  scrollEl?.removeEventListener('scroll', onScroll)
})
</script>

<template>
  <Transition name="page">
    <Button
      v-show="isVisible"
      variant="outline"
      size="icon"
      class="fixed bottom-6 right-6 z-40 h-10 w-10 rounded-full shadow-md"
      :aria-label="t('common.scrollToTop')"
      @click="scrollToTop"
    >
      <ArrowUp class="h-4 w-4" aria-hidden="true" />
    </Button>
  </Transition>
</template>
