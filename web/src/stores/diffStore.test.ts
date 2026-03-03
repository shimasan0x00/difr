import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { useDiffStore } from './diffStore'
import type { DiffFile, DiffMeta, FileContent, FileStats } from '../api/types'
import * as api from '../api/client'

vi.mock('../api/client', () => ({
  toggleReviewedFile: vi.fn().mockResolvedValue({ files: [], reviewed: true }),
  fetchReviewedFiles: vi.fn().mockResolvedValue([]),
  clearReviewedFiles: vi.fn().mockResolvedValue(undefined),
}))

function makeFile(overrides: Partial<DiffFile> = {}): DiffFile {
  return {
    oldPath: 'old.go',
    newPath: 'new.go',
    status: 'modified',
    language: 'go',
    isBinary: false,
    hunks: [],
    stats: { additions: 1, deletions: 0 },
    ...overrides,
  }
}

describe('diffStore', () => {
  beforeEach(() => {
    useDiffStore.setState({
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
    })
  })

  it('setFiles stores files, stats, clears loading, and selects first file', () => {
    const files = [makeFile({ newPath: 'main.go' })]
    const stats: FileStats = { additions: 5, deletions: 2 }

    useDiffStore.getState().setFiles(files, stats)

    const state = useDiffStore.getState()
    expect(state.files).toEqual(files)
    expect(state.stats).toEqual(stats)
    expect(state.loading).toBe(false)
    expect(state.error).toBeNull()
    expect(state.selectedFile).toBe('main.go')
  })

  it('setFiles selects oldPath when newPath is /dev/null', () => {
    const files = [makeFile({ oldPath: 'deleted.go', newPath: '/dev/null' })]

    useDiffStore.getState().setFiles(files, { additions: 0, deletions: 10 })

    expect(useDiffStore.getState().selectedFile).toBe('deleted.go')
  })

  it('setFiles stores meta when provided', () => {
    const files = [makeFile({ newPath: 'main.go' })]
    const stats: FileStats = { additions: 1, deletions: 0 }
    const meta: DiffMeta = { from: 'main', to: 'feature/xyz', mode: 'range' }

    useDiffStore.getState().setFiles(files, stats, meta)

    expect(useDiffStore.getState().meta).toEqual(meta)
  })

  it('setFiles sets meta to null when not provided', () => {
    const files = [makeFile({ newPath: 'main.go' })]

    useDiffStore.getState().setFiles(files, { additions: 1, deletions: 0 })

    expect(useDiffStore.getState().meta).toBeNull()
  })

  it('setMeta updates meta', () => {
    const meta: DiffMeta = { from: 'HEAD~1', to: 'HEAD', mode: 'commit' }

    useDiffStore.getState().setMeta(meta)

    expect(useDiffStore.getState().meta).toEqual(meta)
  })

  it('setTrackedFiles stores tracked files', () => {
    const tracked = ['main.go', 'utils.go', 'README.md']

    useDiffStore.getState().setTrackedFiles(tracked)

    expect(useDiffStore.getState().trackedFiles).toEqual(tracked)
  })

  it('setViewMode toggles between split and unified', () => {
    expect(useDiffStore.getState().viewMode).toBe('split')

    useDiffStore.getState().setViewMode('unified')
    expect(useDiffStore.getState().viewMode).toBe('unified')

    useDiffStore.getState().setViewMode('split')
    expect(useDiffStore.getState().viewMode).toBe('split')
  })

  it('setError stores error and clears loading', () => {
    useDiffStore.getState().setError('connection failed')

    const state = useDiffStore.getState()
    expect(state.error).toBe('connection failed')
    expect(state.loading).toBe(false)
  })

  it('selectFile updates selectedFile', () => {
    useDiffStore.getState().selectFile('main.go')
    expect(useDiffStore.getState().selectedFile).toBe('main.go')

    useDiffStore.getState().selectFile(null)
    expect(useDiffStore.getState().selectedFile).toBeNull()
  })

  it('setSidebarTab switches between changed and all', () => {
    expect(useDiffStore.getState().sidebarTab).toBe('changed')

    useDiffStore.getState().setSidebarTab('all')
    expect(useDiffStore.getState().sidebarTab).toBe('all')

    useDiffStore.getState().setSidebarTab('changed')
    expect(useDiffStore.getState().sidebarTab).toBe('changed')
  })

  it('setFileContent stores file content in cache', () => {
    const fc: FileContent = { path: 'main.go', content: 'package main\n', size: 13 }

    useDiffStore.getState().setFileContent(fc)

    const cache = useDiffStore.getState().fileContentCache
    expect(cache.get('main.go')).toEqual(fc)
  })

  it('setFileContentLoading updates loading state', () => {
    expect(useDiffStore.getState().fileContentLoading).toBe(false)

    useDiffStore.getState().setFileContentLoading(true)
    expect(useDiffStore.getState().fileContentLoading).toBe(true)
  })

  it('setFileContentError stores error', () => {
    useDiffStore.getState().setFileContentError('load failed')

    const state = useDiffStore.getState()
    expect(state.fileContentError).toBe('load failed')
    expect(state.fileContentLoading).toBe(false)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('toggleReviewed', () => {
    it('adds a file to reviewedFiles optimistically', async () => {
      await useDiffStore.getState().toggleReviewed('main.go')

      expect(useDiffStore.getState().reviewedFiles.has('main.go')).toBe(true)
      expect(api.toggleReviewedFile).toHaveBeenCalledWith('main.go')
    })

    it('removes a file from reviewedFiles when toggled again', async () => {
      await useDiffStore.getState().toggleReviewed('main.go')
      await useDiffStore.getState().toggleReviewed('main.go')

      expect(useDiffStore.getState().reviewedFiles.has('main.go')).toBe(false)
    })

    it('tracks multiple files independently', async () => {
      await useDiffStore.getState().toggleReviewed('main.go')
      await useDiffStore.getState().toggleReviewed('utils.go')
      await useDiffStore.getState().toggleReviewed('main.go')

      const { reviewedFiles } = useDiffStore.getState()
      expect(reviewedFiles.has('main.go')).toBe(false)
      expect(reviewedFiles.has('utils.go')).toBe(true)
    })

    it('rolls back on API failure', async () => {
      vi.mocked(api.toggleReviewedFile).mockRejectedValueOnce(new Error('network error'))

      await useDiffStore.getState().toggleReviewed('main.go')

      expect(useDiffStore.getState().reviewedFiles.has('main.go')).toBe(false)
    })
  })

  describe('loadReviewedFiles', () => {
    it('loads reviewed files from API', async () => {
      vi.mocked(api.fetchReviewedFiles).mockResolvedValueOnce(['a.go', 'b.go'])

      await useDiffStore.getState().loadReviewedFiles()

      const { reviewedFiles } = useDiffStore.getState()
      expect(reviewedFiles.has('a.go')).toBe(true)
      expect(reviewedFiles.has('b.go')).toBe(true)
    })

    it('silently ignores load errors', async () => {
      vi.mocked(api.fetchReviewedFiles).mockRejectedValueOnce(new Error('fail'))

      await useDiffStore.getState().loadReviewedFiles()

      expect(useDiffStore.getState().reviewedFiles.size).toBe(0)
    })
  })

  describe('clearReviewedFiles', () => {
    it('clears reviewed files optimistically', async () => {
      useDiffStore.setState({ reviewedFiles: new Set(['a.go', 'b.go']) })

      await useDiffStore.getState().clearReviewedFiles()

      expect(useDiffStore.getState().reviewedFiles.size).toBe(0)
      expect(api.clearReviewedFiles).toHaveBeenCalled()
    })

    it('rolls back on API failure', async () => {
      useDiffStore.setState({ reviewedFiles: new Set(['a.go']) })
      vi.mocked(api.clearReviewedFiles).mockRejectedValueOnce(new Error('fail'))

      await useDiffStore.getState().clearReviewedFiles()

      expect(useDiffStore.getState().reviewedFiles.has('a.go')).toBe(true)
    })
  })
})
