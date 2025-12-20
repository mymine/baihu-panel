<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { useSiteSettings } from '@/composables/useSiteSettings'

const props = defineProps<{
  total: number
  page: number
}>()

const emit = defineEmits<{
  'update:page': [page: number]
}>()

const { pageSize } = useSiteSettings()

const totalPages = computed(() => Math.ceil(props.total / pageSize.value) || 1)

function prevPage() {
  if (props.page > 1) {
    emit('update:page', props.page - 1)
  }
}

function nextPage() {
  if (props.page < totalPages.value) {
    emit('update:page', props.page + 1)
  }
}
</script>

<template>
  <div v-if="total > pageSize" class="flex items-center justify-between px-3 py-2 border-t text-xs text-muted-foreground">
    <span>共 {{ total }} 条</span>
    <div class="flex items-center gap-1">
      <Button variant="ghost" size="icon" class="h-6 w-6" :disabled="page <= 1" @click="prevPage">
        <ChevronLeft class="h-3 w-3" />
      </Button>
      <span class="px-2">{{ page }} / {{ totalPages }}</span>
      <Button variant="ghost" size="icon" class="h-6 w-6" :disabled="page >= totalPages" @click="nextPage">
        <ChevronRight class="h-3 w-3" />
      </Button>
    </div>
  </div>
</template>
