<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-vue-next'
import { getPageNumbers } from '@/composables/usePagination'
import { formatNumber } from '@/lib/utils'

const props = defineProps<{
  currentPage: number
  totalPages: number
  totalItems: number
  pageSize: number
  itemName?: string
}>()

const emit = defineEmits<{
  'update:currentPage': [page: number]
}>()

const { t } = useI18n()

const paginationInfo = computed(() => {
  const start = props.totalItems === 0 ? 0 : (props.currentPage - 1) * props.pageSize + 1
  const end = Math.min(props.currentPage * props.pageSize, props.totalItems)
  return { start, end }
})

const summary = computed(() => t('common.showingRange', {
  start: formatNumber(paginationInfo.value.start),
  end: formatNumber(paginationInfo.value.end),
  total: formatNumber(props.totalItems),
  items: props.itemName || t('common.items'),
}))

const pageNumbers = computed(() => getPageNumbers(props.currentPage, props.totalPages))

function goToPage(page: number | '...') {
  if (page === '...') return
  if (page < 1 || page > props.totalPages || page === props.currentPage) return
  emit('update:currentPage', page)
}
</script>

<template>
  <nav class="flex flex-wrap items-center justify-between gap-3" :aria-label="t('common.pagination')">
    <p class="text-sm text-muted-foreground" aria-live="polite">{{ summary }}</p>
    <div class="flex items-center gap-1">
      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8"
        :disabled="currentPage === 1"
        :aria-label="t('common.firstPage')"
        @click="goToPage(1)"
      >
        <ChevronsLeft class="h-4 w-4" aria-hidden="true" />
      </Button>
      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8"
        :disabled="currentPage === 1"
        :aria-label="t('common.previousPage')"
        @click="goToPage(currentPage - 1)"
      >
        <ChevronLeft class="h-4 w-4" aria-hidden="true" />
      </Button>
      <div class="mx-1 hidden items-center gap-1 sm:flex">
        <template v-for="(page, index) in pageNumbers" :key="index">
          <Button
            v-if="page !== '...'"
            :variant="page === currentPage ? 'default' : 'outline'"
            size="icon"
            class="h-8 w-8"
            :aria-label="t('common.pageNumber', { page })"
            :aria-current="page === currentPage ? 'page' : undefined"
            @click="goToPage(page)"
          >
            {{ page }}
          </Button>
          <span v-else class="px-1 text-muted-foreground" aria-hidden="true">…</span>
        </template>
      </div>
      <span class="mx-1 text-sm text-muted-foreground sm:hidden">{{ currentPage }} / {{ totalPages }}</span>
      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8"
        :disabled="currentPage === totalPages"
        :aria-label="t('common.nextPage')"
        @click="goToPage(currentPage + 1)"
      >
        <ChevronRight class="h-4 w-4" aria-hidden="true" />
      </Button>
      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8"
        :disabled="currentPage === totalPages"
        :aria-label="t('common.lastPage')"
        @click="goToPage(totalPages)"
      >
        <ChevronsRight class="h-4 w-4" aria-hidden="true" />
      </Button>
    </div>
  </nav>
</template>
