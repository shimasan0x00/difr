import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import App from './App'
import { useDiffStore } from './stores/diffStore'
import type { DiffResult } from './api/types'

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
    fetchSpy.mockResolvedValueOnce({
      ok: true,
      json: async () => mockDiffResult,
    } as Response)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('file.go')).toBeInTheDocument()
    })
  })

  it('displays error message on fetch failure', async () => {
    fetchSpy.mockResolvedValueOnce({
      ok: false,
      status: 500,
    } as Response)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText(/Error:/)).toBeInTheDocument()
    })
  })

  it('renders header with application title', async () => {
    fetchSpy.mockResolvedValueOnce({
      ok: true,
      json: async () => mockDiffResult,
    } as Response)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('difr')).toBeInTheDocument()
    })
  })

  it('shows empty state message when no changes exist', async () => {
    fetchSpy.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ files: [], stats: { additions: 0, deletions: 0 } }),
    } as Response)

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('No changes found.')).toBeInTheDocument()
    })
  })
})
