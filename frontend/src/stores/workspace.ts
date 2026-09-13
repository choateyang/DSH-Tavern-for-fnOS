import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { trimSdk } from '../utils/trimSdk'
import { workspaceApi, configApi } from '../api'
import type { WorkspaceData, WorkspaceItem } from '../types/api'

export const useWorkspaceStore = defineStore('workspace', () => {
  const workspaceData = ref<WorkspaceData>({
    items: [],
    archivedSessionIds: []
  })
  const dataLibraryPath = ref('')
  const loading = ref(false)

  const items = computed<WorkspaceItem[]>(() => {
    return (workspaceData.value.items || []).filter(Boolean)
  })

  function updateWorkspaceData(data: Partial<WorkspaceData>) {
    workspaceData.value = {
      items: data.items || [],
      archivedSessionIds: data.archivedSessionIds || []
    }
  }

  async function fetchWorkspaces(): Promise<void> {
    loading.value = true
    try {
      const res = await workspaceApi.getList()
      if (res.success && res.data) {
        updateWorkspaceData(res.data)
      }
    } finally {
      loading.value = false
    }
  }

  async function fetchDataLibraryPath(): Promise<void> {
    const res = await configApi.getConfig()
    if (res.success && res.data?.data_library_path) {
      dataLibraryPath.value = res.data.data_library_path
    }
  }

  async function openDataLibrary(): Promise<{ success: boolean; message?: string }> {
    if (!dataLibraryPath.value) {
      return { success: false, message: '数据目录路径未配置' }
    }
    return trimSdk.openFileManager(dataLibraryPath.value)
  }

  async function openWorkspace(path: string): Promise<{ success: boolean; message?: string }> {
    return trimSdk.openFileManager(path)
  }

  return {
    workspaceData,
    dataLibraryPath,
    loading,
    items,
    updateWorkspaceData,
    fetchWorkspaces,
    fetchDataLibraryPath,
    openDataLibrary,
    openWorkspace
  }
})
