<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AlertTriangle } from 'lucide-vue-next'
import { api } from '@/api'
import { toast } from 'vue-sonner'

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const demoMode = ref(false)

async function loadDemoMode() {
  try {
    const res = await api.settings.getPublicSite()
    demoMode.value = res.demo_mode || false
  } catch {
    // ignore
  }
}

async function changePassword() {
  if (demoMode.value) {
    toast.error('演示模式下不能修改密码')
    return
  }
  if (!oldPassword.value || !newPassword.value) {
    toast.error('请填写完整')
    return
  }
  if (newPassword.value.length < 6) {
    toast.error('新密码至少6位')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    toast.error('两次密码不一致')
    return
  }
  try {
    await api.settings.changePassword({ old_password: oldPassword.value, new_password: newPassword.value })
    toast.success('密码修改成功')
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (e: any) {
    toast.error(e.message || '修改失败')
  }
}

onMounted(loadDemoMode)
</script>

<template>
  <div class="space-y-4">
    <div v-if="demoMode" class="flex items-center gap-2 p-3 rounded-md bg-destructive/10 text-destructive text-sm">
      <AlertTriangle class="h-4 w-4 shrink-0" />
      <span>演示模式下不能修改密码</span>
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-4 items-center gap-2 sm:gap-4">
      <Label class="sm:text-right">原密码</Label>
      <Input v-model="oldPassword" type="password" placeholder="请输入原密码" class="sm:col-span-3" :disabled="demoMode" />
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-4 items-center gap-2 sm:gap-4">
      <Label class="sm:text-right">新密码</Label>
      <Input v-model="newPassword" type="password" placeholder="请输入新密码（至少6位）" class="sm:col-span-3" :disabled="demoMode" />
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-4 items-center gap-2 sm:gap-4">
      <Label class="sm:text-right">确认密码</Label>
      <Input v-model="confirmPassword" type="password" placeholder="请再次输入新密码" class="sm:col-span-3" :disabled="demoMode" />
    </div>
    <div class="flex justify-end pt-2">
      <Button @click="changePassword" :disabled="demoMode">修改密码</Button>
    </div>
  </div>
</template>
