import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect } from 'vitest'
import { DiffViewer } from './DiffViewer'
import type { Comment, DiffFile } from '../../api/types'

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

  it('displays comment count badge in file header when comments exist', () => {
    const fileComments: Comment[] = [
      { id: 'c1', filePath: 'main.go', line: 2, body: 'Fix this', createdAt: '2026-01-01T00:00:00Z' },
      { id: 'c2', filePath: 'main.go', line: 3, body: 'And this', createdAt: '2026-01-01T00:00:00Z' },
    ]
    render(<DiffViewer file={mockFile} viewMode="unified" comments={fileComments} />)
    expect(screen.getByText('2 comments')).toBeInTheDocument()
  })

  it('does not display comment badge when no comments exist', () => {
    render(<DiffViewer file={mockFile} viewMode="unified" comments={[]} />)
    expect(screen.queryByText(/comment/)).not.toBeInTheDocument()
  })

  it('uses singular form for single comment badge', () => {
    const fileComments: Comment[] = [
      { id: 'c1', filePath: 'main.go', line: 2, body: 'Fix this', createdAt: '2026-01-01T00:00:00Z' },
    ]
    render(<DiffViewer file={mockFile} viewMode="unified" comments={fileComments} />)
    expect(screen.getByText('1 comment')).toBeInTheDocument()
  })

  it('collapses diff content when file header is clicked', async () => {
    const user = userEvent.setup()
    render(<DiffViewer file={mockFile} viewMode="unified" />)

    // Initially expanded - diff lines visible
    expect(screen.getByText('func new() {}')).toBeInTheDocument()

    // Click header to collapse
    await user.click(screen.getByRole('button', { name: /toggle/i }))
    expect(screen.queryByText('func new() {}')).not.toBeInTheDocument()
  })

  it('expands diff content when collapsed header is clicked again', async () => {
    const user = userEvent.setup()
    const { container } = render(<DiffViewer file={mockFile} viewMode="unified" />)

    // Collapse
    await user.click(screen.getByRole('button', { name: /toggle/i }))
    expect(container.querySelectorAll('[data-line-type]').length).toBe(0)

    // Expand again
    await user.click(screen.getByRole('button', { name: /toggle/i }))
    expect(container.querySelectorAll('[data-line-type]').length).toBeGreaterThan(0)
  })

  it('sets aria-expanded attribute on file header toggle', () => {
    render(<DiffViewer file={mockFile} viewMode="unified" />)
    const toggle = screen.getByRole('button', { name: /toggle/i })
    expect(toggle).toHaveAttribute('aria-expanded', 'true')
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

  it('does not show comment button on delete-only lines in unified view', () => {
    const deleteOnlyFile: DiffFile = {
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
          newLines: 1,
          header: '',
          lines: [
            { type: 'context', content: 'package main', oldNumber: 1, newNumber: 1 },
            { type: 'delete', content: 'func removed() {}', oldNumber: 2 },
          ],
        },
      ],
      stats: { additions: 0, deletions: 1 },
    }
    render(<DiffViewer file={deleteOnlyFile} viewMode="unified" onAddComment={() => {}} />)
    const deleteLine = screen.getByText('func removed() {}').closest('[data-line-type]')!
    expect(deleteLine.querySelector('[aria-label="Add comment"]')).toBeNull()
  })

  it('does not show comment button on delete-only lines in split view left panel', () => {
    const deleteOnlyFile: DiffFile = {
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
          newLines: 1,
          header: '',
          lines: [
            { type: 'context', content: 'package main', oldNumber: 1, newNumber: 1 },
            { type: 'delete', content: 'func removed() {}', oldNumber: 2 },
          ],
        },
      ],
      stats: { additions: 0, deletions: 1 },
    }
    render(<DiffViewer file={deleteOnlyFile} viewMode="split" onAddComment={() => {}} />)
    const deleteLine = screen.getByText('func removed() {}').closest('[data-line-type]')!
    expect(deleteLine.querySelector('[aria-label="Add comment"]')).toBeNull()
  })

  it('shows comment only on add line, not on delete line with same line number', () => {
    const fileComments: Comment[] = [
      { id: 'c1', filePath: 'main.go', line: 2, body: 'Review this change', createdAt: '2026-01-01T00:00:00Z' },
    ]
    render(<DiffViewer file={mockFile} viewMode="unified" comments={fileComments} />)
    // Comment should appear only once - on the add line (newNumber=2), not the delete line (oldNumber=2)
    expect(screen.getAllByText('Review this change')).toHaveLength(1)
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
