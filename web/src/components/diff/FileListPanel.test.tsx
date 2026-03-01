import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { FileListPanel } from './FileListPanel'
import type { Comment, DiffFile } from '../../api/types'

const mockFiles: DiffFile[] = [
  {
    oldPath: 'main.go',
    newPath: 'main.go',
    status: 'modified',
    language: 'go',
    isBinary: false,
    hunks: [],
    stats: { additions: 3, deletions: 1 },
  },
  {
    oldPath: '/dev/null',
    newPath: 'utils.go',
    status: 'added',
    language: 'go',
    isBinary: false,
    hunks: [],
    stats: { additions: 10, deletions: 0 },
  },
]

const mockCommentsByFile = new Map<string, Comment[]>([
  ['main.go', [
    { id: 'c1', filePath: 'main.go', line: 1, body: 'Fix', createdAt: '2026-01-01T00:00:00Z' },
    { id: 'c2', filePath: 'main.go', line: 2, body: 'Fix2', createdAt: '2026-01-01T00:00:00Z' },
  ]],
])

describe('FileListPanel', () => {
  it('renders file count in header', () => {
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    expect(screen.getByText('Files (2)')).toBeInTheDocument()
  })

  it('renders file paths', () => {
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    expect(screen.getByText('main.go')).toBeInTheDocument()
    expect(screen.getByText('utils.go')).toBeInTheDocument()
  })

  it('renders addition and deletion stats per file', () => {
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    expect(screen.getByText('+3')).toBeInTheDocument()
    expect(screen.getByText('-1')).toBeInTheDocument()
    expect(screen.getByText('+10')).toBeInTheDocument()
  })

  it('renders comment count for files with comments', () => {
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={mockCommentsByFile}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('highlights selected file', () => {
    const { container } = render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile="main.go"
        onSelectFile={vi.fn()}
      />
    )

    const selectedItem = container.querySelector('.border-blue-500')
    expect(selectedItem).toBeInTheDocument()
  })

  it('calls onSelectFile when file is clicked', async () => {
    const user = userEvent.setup()
    const onSelectFile = vi.fn()
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={onSelectFile}
      />
    )

    await user.click(screen.getByText('utils.go'))

    expect(onSelectFile).toHaveBeenCalledWith('utils.go')
  })

  it('collapses file list when header is clicked', async () => {
    const user = userEvent.setup()
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    await user.click(screen.getByRole('button', { name: /toggle file list/i }))

    expect(screen.queryByText('main.go')).not.toBeInTheDocument()
  })

  it('hides file list in controlled mode when expanded is false', () => {
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
        expanded={false}
        onToggle={vi.fn()}
      />
    )

    expect(screen.queryByText('main.go')).not.toBeInTheDocument()
    expect(screen.queryByText('utils.go')).not.toBeInTheDocument()
  })
})
