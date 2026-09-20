<template>
  <el-card shadow="never">
    <template #header>新建投稿</template>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="90px" style="max-width: 840px">
      <el-form-item label="标题" prop="title">
        <el-input v-model="form.title" placeholder="论文标题" />
      </el-form-item>
      <el-form-item label="学科分类" prop="subject">
        <el-select v-model="form.subject" placeholder="请选择学科分类">
          <el-option v-for="s in SUBJECTS" :key="s" :label="SUBJECT_MAP[s]" :value="s" />
        </el-select>
      </el-form-item>
      <el-form-item label="关键词" prop="keywords">
        <el-input v-model="form.keywords" placeholder="多个关键词用英文逗号分隔" />
      </el-form-item>
      <el-form-item label="摘要" prop="abstract">
        <el-input v-model="form.abstract" type="textarea" :rows="4" placeholder="论文摘要（至少 10 字）" />
      </el-form-item>
      <el-form-item label="论文文件" prop="fileKey">
        <el-upload :show-file-list="true" :http-request="doUpload" :file-list="fileList" accept=".pdf,.doc,.docx">
          <el-button type="primary" plain>选择文件（PDF / Word）</el-button>
        </el-upload>
      </el-form-item>
      <el-form-item label="作者" prop="authorsMeta">
        <div style="width: 100%">
          <div v-for="(a, idx) in authors" :key="idx" style="display: flex; gap: 8px; margin-bottom: 8px">
            <el-input v-model="a.name" placeholder="作者姓名" style="width: 180px" />
            <el-input v-model="a.institution" placeholder="单位" style="width: 320px" />
            <el-button type="danger" link :disabled="authors.length === 1" @click="removeAuthor(idx)">
              删除
            </el-button>
          </div>
          <el-button type="primary" link @click="addAuthor">+ 添加作者</el-button>
        </div>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="submit">提交投稿</el-button>
        <el-button @click="router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, UploadRequestOptions } from 'element-plus'
import { SUBJECTS, SUBJECT_MAP } from '../../constants'
import { uploadFile } from '../../api/file'
import { createPaper } from '../../api/paper'

const router = useRouter()
const formRef = ref<FormInstance>()
const submitting = ref(false)
const fileList = ref<Array<{ name: string }>>([])
const authors = ref<Array<{ name: string; institution: string }>>([{ name: '', institution: '' }])
const form = reactive({
  title: '',
  abstract: '',
  keywords: '',
  subject: '',
  fileKey: '',
  fileName: ''
})
const rules: Record<string, unknown> = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  abstract: [{ required: true, min: 10, message: '摘要至少 10 字', trigger: 'blur' }],
  keywords: [{ required: true, message: '请输入关键词', trigger: 'blur' }],
  subject: [{ required: true, message: '请选择学科分类', trigger: 'change' }],
  fileKey: [{ required: true, message: '请上传论文文件', trigger: 'change' }]
}

async function doUpload(options: UploadRequestOptions) {
  const res = await uploadFile(options.file)
  form.fileKey = res.key
  form.fileName = res.name
  fileList.value = [{ name: res.name }]
  ElMessage.success('文件上传成功')
}

function addAuthor() {
  authors.value.push({ name: '', institution: '' })
}

function removeAuthor(idx: number) {
  authors.value.splice(idx, 1)
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  const validAuthors = authors.value.filter((a) => a.name.trim())
  if (!validAuthors.length) {
    ElMessage.warning('请至少填写一位作者')
    return
  }
  submitting.value = true
  try {
    const paper = await createPaper({
      title: form.title,
      abstract: form.abstract,
      keywords: form.keywords,
      subject: form.subject,
      authors_meta: JSON.stringify(validAuthors),
      file_key: form.fileKey,
      file_name: form.fileName
    })
    ElMessage.success('投稿成功，已进入查重与初审队列')
    router.push(`/papers/${paper.id}`)
  } catch {
    // 错误已由拦截器提示
  } finally {
    submitting.value = false
  }
}
</script>
