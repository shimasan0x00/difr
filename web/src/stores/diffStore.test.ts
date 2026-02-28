import { describe, it, expect, beforeEach } from 'vitest'
import { useDiffStore } from './diffStore'
import type { DiffFile, FileStats } from '../api/types'

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
      selectedFile: null,
      viewMode: 'split',
      loading: true,
      error: null,
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
})
