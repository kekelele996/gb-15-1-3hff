import request from '../utils/request'

export interface UploadResult {
  key: string
  name: string
  size: number
}

export function uploadFile(file: File): Promise<UploadResult> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/files/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }) as Promise<UploadResult>
}
