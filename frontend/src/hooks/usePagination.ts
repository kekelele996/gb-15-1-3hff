import { ref } from 'vue'
import type { PageResult } from '../api/types'

export function usePagination<T>(
  fetcher: (params: { page: number; size: number }) => Promise<PageResult<T>>
) {
  const page = ref(1)
  const size = ref(10)
  const total = ref(0)
  const items = ref<T[]>([]) as { value: T[] }
  const loading = ref(false)

  async function load(extra: Record<string, unknown> = {}) {
    loading.value = true
    try {
      const res = await fetcher({ page: page.value, size: size.value, ...extra })
      items.value = res.items
      total.value = res.total
    } finally {
      loading.value = false
    }
  }

  function reset() {
    page.value = 1
    return load()
  }

  return { page, size, total, items, loading, load, reset }
}
