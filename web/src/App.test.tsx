import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import App, { buildReviewContent } from './App'
import { useDiffStore } from './stores/diffStore'
import type { DiffFile, DiffResult } from './api/types'

const mockDiffResult: DiffResult = {
  files: [
    {
      oldPath: 'file.go',
      newPath: 'file.go',
      status: 'modified',
      language: 'go',
      isBinary: false,
      hunks: [
        {
          oldStart: 1,
          oldLines: 2,
          newStart: 1,
          newLines: 2,
          header: '',
          lines: [
            { type: 'delete', content: 'old line', oldNumber: 1 },
            { type: 'add', content: 'new line', newNumber: 1 },
          ],
        },
      ],
      stats: { additions: 1, deletions: 1 },
    },
  ],
  stats: { additions: 1, deletions: 1 },
}

function mockFetchByUrl(diffResult: DiffResult | null, diffStatus = 200) {
  return (url: string | URL | Request) => {
    const urlStr = typeof url === 'string' ? url : url instanceof URL ? url.toString() : url.url
    if (urlStr === '/api/diff') {
      return Promise.resolve({
        ok: diffStatus >= 200 && diffStatus < 300,
        status: diffStatus,
        json: async () => diffResult,
      } as Response)
    }
    if (urlStr === '/api/diff/mode') {
      return Promise.resolve({
        ok: true,
        json: async () => ({ mode: 'split' }),
      } as Response)
    }
    if (urlStr === '/api/comments') {
      return Promise.resolve({
        ok: true,
        json: async () => [],
      } as Response)
    }
    if (urlStr === '/api/diff/tracked-files') {
      return Promise.resolve({
        ok: true,
        json: async () => ({ files: [] }),
      } as Response)
    }
    return Promise.resolve({ ok: true, json: async () => ({}) } as Response)
  }
}

let fetchSpy: ReturnType<typeof vi.spyOn>

beforeEach(() => {
  fetchSpy = vi.spyOn(globalThis, 'fetch')
  useDiffStore.setState({
    files: [],
    stats: { additions: 0, deletions: 0 },
    selectedFile: null,
    viewMode: 'split',
    loading: true,
    error: null,
    sidebarTab: 'changed',
    fileContentCache: new Map(),
    fileContentLoading: false,
    fileContentError: null,
  })
})

afterEach(() => {
  fetchSpy.mockRestore()
})

describe('App', () => {
  it('shows loading state initially', () => {
    fetchSpy.mockImplementation(() => new Promise(() => {}))
    render(<App />)
    expect(screen.getByText('Loading diff...')).toBeInTheDocument()
  })

  it('displays diff files after successful fetch', async () => {
    fetchSpy.mockImplementation(mockFetchByUrl(mockDiffResult) as typeof fetch)

    render(<App />)

    await waitFor(() => {
      // file.go appears in both FileListPanel and DiffViewer
      expect(screen.getAllByText('file.go').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('displays error message on fetch failure', async () => {
    fetchSpy.mockImplementation(mockFetchByUrl(null, 500) as typeof fetch)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText(/Error:/)).toBeInTheDocument()
    })
  })

  it('renders header with application title', async () => {
    fetchSpy.mockImplementation(mockFetchByUrl(mockDiffResult) as typeof fetch)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('difr')).toBeInTheDocument()
    })
  })

  it('displays stats summary bar with file count and additions/deletions', async () => {
    fetchSpy.mockImplementation(mockFetchByUrl(mockDiffResult) as typeof fetch)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('1 file changed')).toBeInTheDocument()
      // +1/-1 appears in both stats summary bar and file header
      expect(screen.getAllByText('+1').length).toBeGreaterThanOrEqual(2)
      expect(screen.getAllByText('-1').length).toBeGreaterThanOrEqual(2)
    })
  })

  it('pluralizes file count in stats summary bar', async () => {
    const multiFileResult: DiffResult = {
      files: [
        mockDiffResult.files[0],
        {
          ...mockDiffResult.files[0],
          oldPath: 'other.go',
          newPath: 'other.go',
        },
      ],
      stats: { additions: 3, deletions: 2 },
    }
    fetchSpy.mockImplementation(mockFetchByUrl(multiFileResult) as typeof fetch)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('2 files changed')).toBeInTheDocument()
    })
  })

  it('shows empty state message when no changes exist', async () => {
    const emptyResult = { files: [], stats: { additions: 0, deletions: 0 } }
    fetchSpy.mockImplementation(mockFetchByUrl(emptyResult) as typeof fetch)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('No changes found.')).toBeInTheDocument()
    })
  })

  it('applies server viewMode on mount', async () => {
    fetchSpy.mockImplementation(((url: string | URL | Request) => {
      const urlStr = typeof url === 'string' ? url : url instanceof URL ? url.toString() : url.url
      if (urlStr === '/api/diff') {
        return Promise.resolve({
          ok: true,
          json: async () => mockDiffResult,
        } as Response)
      }
      if (urlStr === '/api/diff/mode') {
        return Promise.resolve({
          ok: true,
          json: async () => ({ mode: 'unified' }),
        } as Response)
      }
      if (urlStr === '/api/comments') {
        return Promise.resolve({
          ok: true,
          json: async () => [],
        } as Response)
      }
      return Promise.resolve({ ok: true, json: async () => ({}) } as Response)
    }) as typeof fetch)

    render(<App />)

    await waitFor(() => {
      expect(useDiffStore.getState().viewMode).toBe('unified')
    })
  })

  it('shows FileViewer when viewing a non-changed file', async () => {
    fetchSpy.mockImplementation(mockFetchByUrl(mockDiffResult) as typeof fetch)

    render(<App />)

    await waitFor(() => {
      expect(screen.getAllByText('file.go').length).toBeGreaterThanOrEqual(1)
    })

    // Simulate selecting a non-changed file with cached content
    useDiffStore.setState({
      selectedFile: 'readme.md',
      sidebarTab: 'all',
      fileContentCache: new Map([
        ['readme.md', { path: 'readme.md', content: '# Hello\n', size: 8 }],
      ]),
    })

    await waitFor(() => {
      expect(screen.getByText('# Hello')).toBeInTheDocument()
    })
  })
})

describe('buildReviewContent', () => {
  it('includes hunk content with +/- prefixes', () => {
    const files: DiffFile[] = [
      {
        oldPath: 'main.go',
        newPath: 'main.go',
        status: 'modified',
        language: 'go',
        isBinary: false,
        hunks: [
          {
            oldStart: 1,
            oldLines: 2,
            newStart: 1,
            newLines: 2,
            header: '1,2 1,2',
            lines: [
              { type: 'delete', content: 'old line\n', oldNumber: 1 },
              { type: 'add', content: 'new line\n', newNumber: 1 },
              { type: 'context', content: 'same\n', oldNumber: 2, newNumber: 2 },
            ],
          },
        ],
        stats: { additions: 1, deletions: 1 },
      },
    ]

    const result = buildReviewContent(files)

    expect(result).toContain('-old line')
    expect(result).toContain('+new line')
    expect(result).toContain(' same')
  })

  it('includes hunk header', () => {
    const files: DiffFile[] = [
      {
        oldPath: 'main.go',
        newPath: 'main.go',
        status: 'modified',
        language: 'go',
        isBinary: false,
        hunks: [
          {
            oldStart: 10,
            oldLines: 5,
            newStart: 10,
            newLines: 6,
            header: '10,5 10,6',
            lines: [
              { type: 'add', content: 'added\n', newNumber: 10 },
            ],
          },
        ],
        stats: { additions: 1, deletions: 0 },
      },
    ]

    const result = buildReviewContent(files)

    expect(result).toContain('@@ 10,5 10,6 @@')
  })

  it('truncates when content exceeds size limit', () => {
    const longLine = 'x'.repeat(100_000) + '\n'
    const files: DiffFile[] = Array.from({ length: 20 }, (_, i) => ({
      oldPath: `file${i}.go`,
      newPath: `file${i}.go`,
      status: 'modified' as const,
      language: 'go',
      isBinary: false,
      hunks: [
        {
          oldStart: 1,
          oldLines: 1,
          newStart: 1,
          newLines: 1,
          header: '',
          lines: [{ type: 'add' as const, content: longLine, newNumber: 1 }],
        },
      ],
      stats: { additions: 1, deletions: 0 },
    }))

    const result = buildReviewContent(files)

    expect(result.length).toBeLessThanOrEqual(900_000)
    expect(result).toContain('(Diff truncated due to size limit.)')
  })

  it('returns header only for empty files list', () => {
    const result = buildReviewContent([])

    expect(result).toBe('Review the following code changes:\n\n')
  })
})
