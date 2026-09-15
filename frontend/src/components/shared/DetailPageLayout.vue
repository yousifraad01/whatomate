<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ScrollArea } from '@/components/ui/scroll-area'
import PageHeader from './PageHeader.vue'
import ErrorState from './ErrorState.vue'
import { Loader2 } from 'lucide-vue-next'
import type { Component } from 'vue'

defineProps<{
  title: string
  description?: string
  icon?: Component
  iconGradient?: string
  backLink: string
  breadcrumbs?: Array<{ label: string; href?: string }>
  isLoading?: boolean
  isNotFound?: boolean
  notFoundTitle?: string
  notFoundDescription?: string
}>()

const { t } = useI18n()
</script>

<template>
  <div class="flex h-full flex-col bg-background">
    <PageHeader
      :title="title"
      :description="description"
      :icon="icon"
      :back-link="backLink"
      :breadcrumbs="breadcrumbs"
    >
      <template #actions>
        <slot name="actions" />
      </template>
    </PageHeader>

    <!-- Loading -->
    <div v-if="isLoading" class="flex flex-1 items-center justify-center" role="status" aria-live="polite">
      <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" aria-hidden="true" />
      <span class="sr-only">{{ t('common.loading') }}</span>
    </div>

    <!-- Not found -->
    <ErrorState
      v-else-if="isNotFound"
      :title="notFoundTitle || t('common.notFoundTitle')"
      :description="notFoundDescription || t('common.notFoundDesc')"
      class="flex-1"
    />

    <!-- Content -->
    <ScrollArea v-else class="flex-1">
      <div class="p-4 sm:p-6">
        <div class="mx-auto grid max-w-6xl grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="space-y-6 lg:col-span-2">
            <slot />
          </div>
          <div class="space-y-6">
            <slot name="sidebar" />
          </div>
        </div>
      </div>
    </ScrollArea>
  </div>
</template>
