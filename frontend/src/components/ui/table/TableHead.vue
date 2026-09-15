<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
  scope?: "col" | "row"
}>()
</script>

<template>
  <th
    :scope="props.scope ?? 'col'"
    :class="cn('h-10 px-3 text-left align-middle text-xs font-medium uppercase tracking-wide text-muted-foreground [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-0.5', props.class)"
  >
    <slot />
  </th>
</template>
