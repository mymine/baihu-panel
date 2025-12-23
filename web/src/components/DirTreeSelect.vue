<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Folder, ChevronRight, ChevronDown, FolderOpen } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { api, type FileNode } from '@/api'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const loading = ref(false)
const fileTree = ref<FileNode[]>([])
const expandedDirs = ref<Set<string>>(new Set())

// 只保留目录节点
function filterDirs(nodes: FileNode[]): FileNode[] {
  return nodes
    .filter(n => n.isDir)
    .map(n => ({
      ...n,
      children: n.children ? filterDirs(n.children) : undefined
    }))
}

const dirTree = computed(() => filterDirs(fileTree.value))

async function loadTree() {
  if (fileTree.value.length > 0) return
  loading.value = true
  try {
    fileTree.value = await api.files.tree()
  } catch {
    fileTree.value = []
  } finally {
    loading.value = false
  }
}

function toggleDir(path: string) {
  if (expandedDirs.value.has(path)) {
    expandedDirs.value.delete(path)
  } else {
    expandedDirs.value.add(path)
  }
}

function selectDir(path: string) {
  emit('update:modelValue', path)
  open.value = false
}

function selectRoot() {
  emit('update:modelValue', '')
  open.value = false
}

watch(open, (val) => {
  if (val) loadTree()
})

const displayValue = computed(() => {
  if (!props.modelValue) return props.placeholder || 'scripts (默认)'
  return props.modelValue
})
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button variant="outline" class="w-full justify-start font-mono text-sm h-9">
        <Folder class="h-4 w-4 mr-2 text-yellow-500 shrink-0" />
        <span class="truncate">{{ displayValue }}</span>
      </Button>
    </PopoverTrigger>
    <PopoverContent class="w-[300px] p-2" align="start">
      <div class="text-xs text-muted-foreground mb-2">选择工作目录</div>
      <div class="max-h-[240px] overflow-y-auto">
        <!-- 根目录选项 -->
        <div
          :class="[
            'flex items-center gap-1.5 py-1 px-2 rounded cursor-pointer text-sm',
            !modelValue ? 'bg-primary/10 text-primary' : 'hover:bg-muted'
          ]"
          @click="selectRoot"
        >
          <FolderOpen class="h-4 w-4 text-yellow-500" />
          <span>scripts (默认)</span>
        </div>
        
        <!-- 目录树 -->
        <DirTreeNode
          v-for="node in dirTree"
          :key="node.path"
          :node="node"
          :expanded-dirs="expandedDirs"
          :selected-path="modelValue"
          @toggle="toggleDir"
          @select="selectDir"
        />
        
        <div v-if="loading" class="text-xs text-muted-foreground text-center py-4">
          加载中...
        </div>
        <div v-else-if="dirTree.length === 0" class="text-xs text-muted-foreground text-center py-2">
          暂无子目录
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>

<script lang="ts">
import { defineComponent, h } from 'vue'

// 内联递归组件
const DirTreeNode = defineComponent({
  name: 'DirTreeNode',
  props: {
    node: { type: Object as () => FileNode, required: true },
    expandedDirs: { type: Object as () => Set<string>, required: true },
    selectedPath: { type: String, default: '' },
    depth: { type: Number, default: 0 }
  },
  emits: ['toggle', 'select'],
  setup(props, { emit }) {
    const isExpanded = computed(() => props.expandedDirs.has(props.node.path))
    const isSelected = computed(() => props.selectedPath === props.node.path)
    
    return () => h('div', [
      h('div', {
        class: [
          'flex items-center gap-1 py-1 px-2 rounded cursor-pointer text-sm',
          isSelected.value ? 'bg-primary/10 text-primary' : 'hover:bg-muted'
        ],
        style: { paddingLeft: (props.depth * 12 + 8) + 'px' },
        onClick: () => emit('select', props.node.path)
      }, [
        h('span', {
          class: 'shrink-0 cursor-pointer',
          onClick: (e: Event) => {
            e.stopPropagation()
            emit('toggle', props.node.path)
          }
        }, [
          isExpanded.value
            ? h(ChevronDown, { class: 'h-3 w-3' })
            : h(ChevronRight, { class: 'h-3 w-3' })
        ]),
        h(Folder, { class: 'h-4 w-4 text-yellow-500 shrink-0' }),
        h('span', { class: 'truncate' }, props.node.name)
      ]),
      isExpanded.value && props.node.children?.length
        ? props.node.children.map(child =>
            h(DirTreeNode, {
              key: child.path,
              node: child,
              expandedDirs: props.expandedDirs,
              selectedPath: props.selectedPath,
              depth: props.depth + 1,
              onToggle: (path: string) => emit('toggle', path),
              onSelect: (path: string) => emit('select', path)
            })
          )
        : null
    ])
  }
})

export { DirTreeNode }
</script>
