import { create } from 'zustand'
import type { DiffFile, FileStats } from '../api/types'

type ViewMode = 'split' | 'unified'

interface DiffState {
  files: DiffFile[]
  stats: FileStats
  selectedFile: string | null
  viewMode: ViewMode
  loading: boolean
  error: string | null
  setFiles: (files: DiffFile[], stats: FileStats) => void
  selectFile: (path: string | null) => void
  setViewMode: (mode: ViewMode) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

export const useDiffStore = create<DiffState>((set) => ({
  files: [],
  stats: { additions: 0, deletions: 0 },
  selectedFile: null,
  viewMode: 'split',
  loading: true,
  error: null,
  setFiles: (files, stats) =>
    set({
      files,
      stats,
      loading: false,
      error: null,
      selectedFile: files.length > 0
        ? (files[0].newPath && files[0].newPath !== '/dev/null'
          ? files[0].newPath
          : files[0].oldPath)
        : null,
    }),
  selectFile: (path) => set({ selectedFile: path }),
  setViewMode: (mode) => set({ viewMode: mode }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error, loading: false }),
}))
