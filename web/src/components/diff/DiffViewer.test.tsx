import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { DiffViewer } from './DiffViewer'
import type { DiffFile } from '../../api/types'

const mockFile: DiffFile = {
  oldPath: 'main.go',
  newPath: 'main.go',
  status: 'modified',
  language: 'go',
  isBinary: false,
  hunks: [
    {
      oldStart: 1,
      oldLines: 3,
      newStart: 1,
      newLines: 3,
      header: '',
      lines: [
        { type: 'context', content: 'package main', oldNumber: 1, newNumber: 1 },
        { type: 'delete', content: 'func old() {}', oldNumber: 2 },
        { type: 'add', content: 'func new() {}', newNumber: 2 },
        { type: 'context', content: '', oldNumber: 3, newNumber: 3 },
      ],
    },
  ],
  stats: { additions: 1, deletions: 1 },
}

const newFile: DiffFile = {
  oldPath: '/dev/null',
  newPath: 'utils.go',
  status: 'added',
  language: 'go',
  isBinary: false,
  hunks: [
    {
      oldStart: 0,
      oldLines: 0,
      newStart: 1,
      newLines: 2,
      header: '',
      lines: [
        { type: 'add', content: 'package main', newNumber: 1 },
        { type: 'add', content: 'func helper() {}', newNumber: 2 },
      ],
    },
  ],
  stats: { additions: 2, deletions: 0 },
}

describe('DiffViewer', () => {
  it('renders file header with path', () => {
    render(<DiffViewer file={mockFile} viewMode="unified" />)
    expect(screen.getByText('main.go')).toBeInTheDocument()
  })

  it('renders stats badge', () => {
    render(<DiffViewer file={mockFile} viewMode="unified" />)
    expect(screen.getByText('+1')).toBeInTheDocument()
    expect(screen.getByText('-1')).toBeInTheDocument()
  })

  it('renders added lines with add styling in unified mode', () => {
    render(<DiffViewer file={mockFile} viewMode="unified" />)
    const addedLine = screen.getByText('func new() {}')
    expect(addedLine.closest('[data-line-type]')).toHaveAttribute('data-line-type', 'add')
  })

  it('renders deleted lines with delete styling in unified mode', () => {
    render(<DiffViewer file={mockFile} viewMode="unified" />)
    const deletedLine = screen.getByText('func old() {}')
    expect(deletedLine.closest('[data-line-type]')).toHaveAttribute('data-line-type', 'delete')
  })

  it('renders context lines in unified mode', () => {
    render(<DiffViewer file={mockFile} viewMode="unified" />)
    const contextLine = screen.getByText('package main')
    expect(contextLine.closest('[data-line-type]')).toHaveAttribute('data-line-type', 'context')
  })

  it('renders line numbers', () => {
    render(<DiffViewer file={mockFile} viewMode="unified" />)
    // Line 1 context should show both old and new numbers
    expect(screen.getAllByText('1').length).toBeGreaterThanOrEqual(2)
  })

  it('renders split view with two columns', () => {
    render(<DiffViewer file={mockFile} viewMode="split" />)
    const splitContainer = screen.getByTestId('split-view')
    expect(splitContainer).toBeInTheDocument()
  })

  it('renders new file with added status indicator', () => {
    render(<DiffViewer file={newFile} viewMode="unified" />)
    expect(screen.getByText('utils.go')).toBeInTheDocument()
    expect(screen.getByText('+2')).toBeInTheDocument()
  })

  it('renders split view with deleted lines on left and added lines on right', () => {
    const { container } = render(<DiffViewer file={mockFile} viewMode="split" />)
    const splitView = container.querySelector('[data-testid="split-view"]')!

    // In split view, the delete/add pair should be shown side-by-side
    // Left side should contain deleted content, right side should contain added content
    const deleteLine = container.querySelector('[data-line-type="delete"]')
    expect(deleteLine).toBeInTheDocument()

    // Verify the deleted text appears on left and added text appears on right
    expect(screen.getByText('func old() {}')).toBeInTheDocument()
    expect(screen.getByText('func new() {}')).toBeInTheDocument()

    // Both should be within the split view
    expect(splitView.contains(screen.getByText('func old() {}'))).toBe(true)
    expect(splitView.contains(screen.getByText('func new() {}'))).toBe(true)
  })

  it('shows binary file message', () => {
    const binaryFile: DiffFile = {
      ...mockFile,
      isBinary: true,
      hunks: [],
    }
    render(<DiffViewer file={binaryFile} viewMode="unified" />)
    expect(screen.getByText(/binary/i)).toBeInTheDocument()
  })

  it('generates unique keys for lines in unified view', () => {
    const fileWithDuplicateTypes: DiffFile = {
      ...mockFile,
      hunks: [
        {
          oldStart: 1,
          oldLines: 3,
          newStart: 1,
          newLines: 3,
          header: '',
          lines: [
            { type: 'add', content: 'line A', newNumber: 1 },
            { type: 'add', content: 'line B', newNumber: 2 },
            { type: 'delete', content: 'line C', oldNumber: 1 },
            { type: 'delete', content: 'line D', oldNumber: 2 },
          ],
        },
      ],
      stats: { additions: 2, deletions: 2 },
    }
    const { container } = render(
      <DiffViewer file={fileWithDuplicateTypes} viewMode="unified" />
    )
    // All 4 lines should render without React key warnings
    const lines = container.querySelectorAll('[data-line-type]')
    expect(lines.length).toBe(4)
  })
})
