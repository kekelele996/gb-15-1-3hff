<template>
  <el-card shadow="never" style="max-width: 840px">
    <template #header>提交修改稿</template>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
      <el-form-item label="论文文件" prop="fileKey">
        <el-upload :show-file-list="true" :http-request="doUpload" :file-list="fileList" accept=".pdf,.doc,.docx">
          <el-button type="primary" plain>选择修改稿（PDF / Word）</el-button>
        </el-upload>
      </el-form-item>
      <el-form-item label="修改说明" prop="responseLetter">
        <el-input
          v-model="form.responseLetter"
          type="textarea"
          :rows="8"
          placeholder="请逐条回复审稿意见，说明修改内容"
        />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="submit">提交修改稿</el-button>
        <el-button @click="router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, UploadRequestOptions } from 'element-plus'
import { uploadFile } from '../../api/file'
import { revisePaper } from '../../api/paper'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const submitting = ref(false)
const fileList = ref<Array<{ name: string }>>([])
const form = reactive({ fileKey: '', fileName: '', responseLetter: '' })
const rules: Record<string, unknown> = {
  fileKey: [{ required: true, message: '请上传修改稿', trigger: 'change' }],
  responseLetter: [{ required: true, min: 10, message: '修改说明至少 10 字', trigger: 'blur' }]
}

async function doUpload(options: UploadRequestOptions) {
  const res = await uploadFile(options.file)
  form.fileKey = res.key
  form.fileName = res.name
  fileList.value = [{ name: res.name }]
  ElMessage.success('文件上传成功')
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const paper = await revisePaper(route.params.id as string, {
      file_key: form.fileKey,
      file_name: form.fileName,
      response_letter: form.responseLetter
    })
    ElMessage.success('修改稿已提交，进入下一轮评审')
    router.push(`/papers/${paper.id}`)
  } catch {
    // 拦截器已提示
  } finally {
    submitting.value = false
  }
}
</script>
