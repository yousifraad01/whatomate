<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { ArrowLeft } from 'lucide-vue-next'
import type { Component } from 'vue'

defineProps<{
  title: string
  description?: string
  /** Alias of description, kept for existing call sites */
  subtitle?: string
  icon?: Component
  /** @deprecated Decorative icon tiles were removed; the prop is accepted so existing call sites keep compiling. */
  iconGradient?: string
  backLink?: string
  breadcrumbs?: Array<{ label: string; href?: string }>
}>()

const { t } = useI18n()
</script>

<template>
  <header class="shrink-0 border-b border-border bg-background">
    <div class="flex min-h-14 flex-wrap items-center gap-x-3 gap-y-2 px-4 py-2 sm:px-6">
      <Button
        v-if="backLink"
        as-child
        variant="ghost"
        size="icon"
        class="-ml-2 h-8 w-8 shrink-0"
      >
        <RouterLink :to="backLink" :aria-label="t('common.back')">
          <ArrowLeft class="h-4 w-4" aria-hidden="true" />
        </RouterLink>
      </Button>
      <component
        v-if="icon"
        :is="icon"
        class="hidden h-5 w-5 shrink-0 text-muted-foreground sm:block"
        aria-hidden="true"
      />
      <div class="min-w-0 flex-1">
        <h1 class="truncate text-lg font-semibold leading-tight text-foreground">{{ title }}</h1>
        <Breadcrumb v-if="breadcrumbs?.length" class="mt-0.5">
          <BreadcrumbList class="text-xs">
            <template v-for="(crumb, index) in breadcrumbs" :key="index">
              <BreadcrumbItem>
                <BreadcrumbLink v-if="crumb.href" :href="crumb.href">
                  {{ crumb.label }}
                </BreadcrumbLink>
                <BreadcrumbPage v-else>{{ crumb.label }}</BreadcrumbPage>
              </BreadcrumbItem>
              <BreadcrumbSeparator v-if="index < breadcrumbs.length - 1" />
            </template>
          </BreadcrumbList>
        </Breadcrumb>
        <p v-else-if="description || subtitle" class="mt-0.5 line-clamp-2 text-sm text-muted-foreground sm:line-clamp-none sm:truncate">
          {{ description || subtitle }}
        </p>
      </div>
      <div v-if="$slots.actions" class="flex w-full flex-wrap items-center gap-2 lg:w-auto">
        <slot name="actions" />
      </div>
    </div>
  </header>
</template>
