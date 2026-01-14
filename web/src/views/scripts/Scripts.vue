<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { RefreshCw, FolderPlus, FilePlus, Save } from 'lucide-vue-next'
import { api, type FileNode } from '@/api'
import { VueMonacoEditor } from '@guolao/vue-monaco-editor'
import FileTreeNode from '@/components/FileTreeNode.vue'
import { toast } from 'vue-sonner'

const route = useRoute()
const router = useRouter()

const fileTree = ref<FileNode[]>([])
const expandedDirs = ref<Set<string>>(new Set())
const selectedFile = ref<string | null>(null)
const selectedDir = ref<string | null>(null)  // 当前选中的文件夹
const fileContent = ref('')
const originalContent = ref('')
const loading = ref(false)
const saving = ref(false)

const showCreateDialog = ref(false)
const createType = ref<'file' | 'folder'>('file')
const createName = ref('')
const showDeleteDialog = ref(false)
const deletePath = ref<string | null>(null)
const showUnsavedDialog = ref(false)
const pendingNode = ref<FileNode | null>(null)

const hasChanges = computed(() => fileContent.value !== originalContent.value)

const editorLanguage = computed(() => {
  if (!selectedFile.value) return 'plaintext'
  const name = selectedFile.value.toLowerCase()
  if (name.endsWith('.ts')) return 'typescript'
  if (name.endsWith('.js')) return 'javascript'
  if (name.endsWith('.py')) return 'python'
  if (name.endsWith('.sh')) return 'shell'
  if (name.endsWith('.json')) return 'json'
  if (name.endsWith('.yaml') || name.endsWith('.yml')) return 'yaml'
  if (name.endsWith('.go')) return 'go'
  return 'plaintext'
})

// 更新 URL - 每次清空重建
function updateUrl() {
  const query: Record<string, string> = {}
  if (selectedFile.value) query.file = selectedFile.value
  if (selectedDir.value) query.dir = selectedDir.value
  if (expandedDirs.value.size > 0) query.dirs = Array.from(expandedDirs.value).join(',')
  router.replace({ path: route.path, query })
}

// 展开文件所在的所有父目录
function expandParentDirs(filePath: string) {
  const parts = filePath.split('/')
  let current: string = ''
  for (let i = 0; i < parts.length - 1; i++) {
    current = current ? `${current}/${parts[i]}` : parts[i] ?? ''
    expandedDirs.value.add(current)
  }
}

async function loadTree() {
  loading.value = true
  try {
    fileTree.value = await api.files.tree()
    
    // 仅在首次加载时从 URL 恢复状态
    if (expandedDirs.value.size === 0 && selectedFile.value === null && selectedDir.value === null) {
      // 从 URL 恢复展开的目录
      const dirsParam = route.query.dirs
      if (dirsParam && typeof dirsParam === 'string') {
        dirsParam.split(',').forEach(dir => expandedDirs.value.add(dir))
      }
      
      // 从 URL 恢复选中的文件夹
      const dirParam = route.query.dir
      if (dirParam && typeof dirParam === 'string') {
        selectedDir.value = dirParam
        expandedDirs.value.add(dirParam)
      }
      
      // 从 URL 加载文件
      const fileParam = route.query.file
      if (fileParam && typeof fileParam === 'string') {
        expandParentDirs(fileParam)
        await loadFileContent(fileParam)
      }
    }
  } catch {
    fileTree.value = []
  } finally {
    loading.value = false
  }
}

async function loadFileContent(path: string) {
  try {
    const res = await api.files.getContent(path)
    selectedFile.value = path
    fileContent.value = res.content
    originalContent.value = res.content
  } catch {
    selectedFile.value = null
    fileContent.value = ''
    originalContent.value = ''
  }
}

function toggleDir(path: string) {
  if (expandedDirs.value.has(path)) {
    expandedDirs.value.delete(path)
  } else {
    expandedDirs.value.add(path)
  }
  // 点击文件夹时不改变文件选择状态，只更新展开状态
  updateUrl()
}

async function handleSelect(node: FileNode) {
  if (node.isDir) {
    selectedDir.value = node.path
    selectedFile.value = null
    fileContent.value = ''
    originalContent.value = ''
    toggleDir(node.path)
    return
  }
  
  if (hasChanges.value) {
    pendingNode.value = node
    showUnsavedDialog.value = true
    return
  }
  
  await selectFile(node)
}

async function selectFile(node: FileNode) {
  selectedDir.value = null
  loading.value = true
  try {
    await loadFileContent(node.path)
    expandParentDirs(node.path)
    updateUrl()
  } finally {
    loading.value = false
  }
}

async function confirmSwitchFile() {
  showUnsavedDialog.value = false
  if (pendingNode.value) {
    await selectFile(pendingNode.value)
    pendingNode.value = null
  }
}

async function saveFile() {
  if (!selectedFile.value) return
  saving.value = true
  try {
    await api.files.saveContent(selectedFile.value, fileContent.value)
    originalContent.value = fileContent.value
    toast.success('文件已保存')
  } catch {
    toast.error('保存失败')
  } finally {
    saving.value = false
  }
}

function openCreateDialog(type: 'file' | 'folder') {
  createType.value = type
  createName.value = ''
  showCreateDialog.value = true
}

// 计算完整路径（选中文件夹 + 文件名）
const createFullPath = computed(() => {
  if (!createName.value) return ''
  return selectedDir.value ? `${selectedDir.value}/${createName.value}` : createName.value
})

async function createItem() {
  if (!createName.value) return
  const fullPath = createFullPath.value
  const currentSelectedDir = selectedDir.value
  try {
    await api.files.create(fullPath, createType.value === 'folder')
    showCreateDialog.value = false
    toast.success(createType.value === 'file' ? '文件已创建' : '文件夹已创建')
    if (currentSelectedDir) {
      expandedDirs.value.add(currentSelectedDir)
    }
    await loadTree()
    selectedDir.value = currentSelectedDir
    if (createType.value === 'file') {
      await loadFileContent(fullPath)
      selectedDir.value = null
      updateUrl()
    }
  } catch { toast.error('创建失败') }
}

function confirmDeleteFile(path: string) {
  deletePath.value = path
  showDeleteDialog.value = true
}

async function handleDelete() {
  if (!deletePath.value) return
  const path = deletePath.value
  const currentSelectedDir = selectedDir.value
  try {
    await api.files.delete(path)
    toast.success('已删除')
    if (selectedFile.value === path) {
      selectedFile.value = null
      fileContent.value = ''
      originalContent.value = ''
      updateUrl()
    }
    if (selectedDir.value === path) {
      selectedDir.value = null
    }
    await loadTree()
    if (currentSelectedDir && currentSelectedDir !== path) {
      selectedDir.value = currentSelectedDir
    }
  } catch { toast.error('删除失败') }
  showDeleteDialog.value = false
  deletePath.value = null
}

onMounted(loadTree)
</script>

<template>
  <div class="flex flex-col lg:flex-row h-[calc(100vh-2rem)] gap-3">
    <!-- File Tree -->
    <div class="w-full lg:w-56 flex-shrink-0 border rounded-lg bg-card flex flex-col max-h-[200px] lg:max-h-none">
      <div class="p-2 border-b flex items-center justify-between">
        <span class="text-xs font-medium">脚本文件</span>
        <div class="flex gap-0.5">
          <Button variant="ghost" size="icon" class="h-6 w-6" title="新建文件" @click="openCreateDialog('file')">
            <FilePlus class="h-3 w-3" />
          </Button>
          <Button variant="ghost" size="icon" class="h-6 w-6" title="新建文件夹" @click="openCreateDialog('folder')">
            <FolderPlus class="h-3 w-3" />
          </Button>
          <Button variant="ghost" size="icon" class="h-6 w-6" title="刷新" @click="loadTree">
            <RefreshCw class="h-3 w-3" :class="{ 'animate-spin': loading }" />
          </Button>
        </div>
      </div>
      
      <div class="flex-1 overflow-auto p-1">
        <div v-if="fileTree.length === 0" class="text-xs text-muted-foreground text-center py-4">
          暂无文件
        </div>
        <FileTreeNode
          v-for="node in fileTree"
          :key="node.path"
          :node="node"
          :expanded-dirs="expandedDirs"
          :selected-path="selectedFile || selectedDir"
          @select="handleSelect"
          @delete="confirmDeleteFile"
        />
      </div>
    </div>

    <!-- Editor -->
    <div class="flex-1 border rounded-lg bg-card flex flex-col overflow-hidden min-h-[300px]">
      <div class="p-2 border-b flex items-center justify-between">
        <div class="flex items-center gap-2 min-w-0">
          <span class="text-xs font-medium truncate">{{ selectedFile || '未选择文件' }}</span>
          <span v-if="hasChanges" class="text-xs text-orange-500 shrink-0">● 未保存</span>
        </div>
        <Button 
          v-if="selectedFile" 
          size="sm" 
          class="h-6 text-xs gap-1 shrink-0" 
          :disabled="!hasChanges || saving"
          @click="saveFile"
        >
          <Save class="h-3 w-3" />
          <span class="hidden sm:inline">{{ saving ? '保存中...' : '保存' }}</span>
        </Button>
      </div>
      
      <div class="flex-1">
        <VueMonacoEditor
          v-if="selectedFile"
          v-model:value="fileContent"
          :language="editorLanguage"
          theme="vs-dark"
          :options="{
            minimap: { enabled: false },
            fontSize: 13,
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            wordWrap: 'on',
            files: {
              eol: '\n'
            }
          }"
          style="height: 100%"
        />
        <div v-else class="h-full flex items-center justify-center text-muted-foreground text-sm">
          选择一个文件开始编辑
        </div>
      </div>
    </div>

    <!-- Create Dialog -->
    <Dialog v-model:open="showCreateDialog">
      <DialogContent class="max-w-xs">
        <DialogHeader>
          <DialogTitle class="text-sm">
            {{ createType === 'file' ? '新建文件' : '新建文件夹' }}
          </DialogTitle>
        </DialogHeader>
        <div class="py-2 space-y-2">
          <div v-if="selectedDir" class="text-xs text-muted-foreground">
            位置: {{ selectedDir }}/
          </div>
          <Input 
            v-model="createName" 
            class="h-8 text-xs" 
            :placeholder="createType === 'file' ? 'example.js' : 'folder-name'"
            @keyup.enter="createItem"
          />
          <div v-if="createName" class="text-xs text-muted-foreground">
            完整路径: {{ createFullPath }}
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" size="sm" class="h-7 text-xs" @click="showCreateDialog = false">取消</Button>
          <Button size="sm" class="h-7 text-xs" @click="createItem">创建</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <AlertDialog v-model:open="showDeleteDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle class="text-sm">确认删除</AlertDialogTitle>
          <AlertDialogDescription class="text-xs">确定要删除 {{ deletePath }} 吗？此操作无法撤销。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel class="h-7 text-xs">取消</AlertDialogCancel>
          <AlertDialogAction class="h-7 text-xs bg-destructive text-white hover:bg-destructive/90" @click="handleDelete">删除</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <AlertDialog v-model:open="showUnsavedDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle class="text-sm">未保存的更改</AlertDialogTitle>
          <AlertDialogDescription class="text-xs">当前文件有未保存的更改，确定要切换文件吗？</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel class="h-7 text-xs">取消</AlertDialogCancel>
          <AlertDialogAction class="h-7 text-xs" @click="confirmSwitchFile">确定切换</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
