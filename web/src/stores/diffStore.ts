import { create } from 'zustand'
import type { DiffFile, DiffMeta, FileContent, FileStats } from '../api/types'
import * as api from '../api/client'

type ViewMode = 'split' | 'unified'
type SidebarTab = 'changed' | 'all'

interface DiffState {
  files: DiffFile[]
  stats: FileStats
  meta: DiffMeta | null
  trackedFiles: string[]
  selectedFile: string | null
  viewMode: ViewMode
  loading: boolean
  error: string | null
  sidebarTab: SidebarTab
  fileContentCache: Map<string, FileContent>
  fileContentLoading: boolean
  fileContentError: string | null
  reviewedFiles: Set<string>
  toggleReviewed: (path: string) => void
  loadReviewedFiles: () => Promise<void>
  clearReviewedFiles: () => Promise<void>
  setFiles: (files: DiffFile[], stats: FileStats, meta?: DiffMeta) => void
  setMeta: (meta: DiffMeta) => void
  setTrackedFiles: (files: string[]) => void
  selectFile: (path: string | null) => void
  setViewMode: (mode: ViewMode) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setSidebarTab: (tab: SidebarTab) => void
  setFileContent: (fc: FileContent) => void
  setFileContentLoading: (loading: boolean) => void
  setFileContentError: (error: string | null) => void
}

export const useDiffStore = create<DiffState>((set, get) => ({
  files: [],
  stats: { additions: 0, deletions: 0 },
  meta: null,
  trackedFiles: [],
  selectedFile: null,
  viewMode: 'split',
  loading: true,
  error: null,
  sidebarTab: 'changed',
  fileContentCache: new Map(),
  fileContentLoading: false,
  fileContentError: null,
  reviewedFiles: new Set(),
  toggleReviewed: async (path) => {
    const prev = get().reviewedFiles
    const next = new Set(prev)
    if (next.has(path)) {
      next.delete(path)
    } else {
      next.add(path)
    }
    set({ reviewedFiles: next })
    try {
      await api.toggleReviewedFile(path)
    } catch {
      set({ reviewedFiles: prev })
    }
  },
  loadReviewedFiles: async () => {
    try {
      const files = await api.fetchReviewedFiles()
      set({ reviewedFiles: new Set(files) })
    } catch {
      // Silently ignore load errors
    }
  },
  clearReviewedFiles: async () => {
    const prev = get().reviewedFiles
    set({ reviewedFiles: new Set() })
    try {
      await api.clearReviewedFiles()
    } catch {
      set({ reviewedFiles: prev })
    }
  },
  setFiles: (files, stats, meta) =>
    set({
      files,
      stats,
      meta: meta ?? null,
      loading: false,
      error: null,
      selectedFile: files.length > 0
        ? (files[0].newPath && files[0].newPath !== '/dev/null'
          ? files[0].newPath
          : files[0].oldPath)
        : null,
    }),
  setMeta: (meta) => set({ meta }),
  setTrackedFiles: (files) => set({ trackedFiles: files }),
  selectFile: (path) => set({ selectedFile: path }),
  setViewMode: (mode) => set({ viewMode: mode }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error, loading: false }),
  setSidebarTab: (tab) => set({ sidebarTab: tab }),
  setFileContent: (fc) => {
    const cache = new Map(get().fileContentCache)
    cache.set(fc.path, fc)
    set({ fileContentCache: cache, fileContentLoading: false, fileContentError: null })
  },
  setFileContentLoading: (loading) => set({ fileContentLoading: loading }),
  setFileContentError: (error) => set({ fileContentError: error, fileContentLoading: false }),
}))
