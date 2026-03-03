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
        activeTab="changed"
        onTabChange={vi.fn()}
      />
    )

    expect(screen.getByText('Files (2)')).toBeInTheDocument()
  })

  it('renders file paths in Changed tab', () => {
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
        activeTab="changed"
        onTabChange={vi.fn()}
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
        activeTab="changed"
        onTabChange={vi.fn()}
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
        activeTab="changed"
        onTabChange={vi.fn()}
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
        activeTab="changed"
        onTabChange={vi.fn()}
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
        activeTab="changed"
        onTabChange={vi.fn()}
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
        activeTab="changed"
        onTabChange={vi.fn()}
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
        activeTab="changed"
        onTabChange={vi.fn()}
      />
    )

    expect(screen.queryByText('main.go')).not.toBeInTheDocument()
    expect(screen.queryByText('utils.go')).not.toBeInTheDocument()
  })

  it('shows Changed and All Files tabs when trackedFiles provided', () => {
    render(
      <FileListPanel
        files={mockFiles}
        trackedFiles={['main.go', 'utils.go', 'readme.md']}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
        activeTab="changed"
        onTabChange={vi.fn()}
      />
    )

    expect(screen.getByRole('tab', { name: /changed/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /all files/i })).toBeInTheDocument()
  })

  it('calls onTabChange when All Files tab is clicked', async () => {
    const user = userEvent.setup()
    const onTabChange = vi.fn()
    render(
      <FileListPanel
        files={mockFiles}
        trackedFiles={['main.go', 'utils.go', 'readme.md']}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
        activeTab="changed"
        onTabChange={onTabChange}
      />
    )

    await user.click(screen.getByRole('tab', { name: /all files/i }))

    expect(onTabChange).toHaveBeenCalledWith('all')
  })

  it('shows directory tree when All Files tab is active', () => {
    render(
      <FileListPanel
        files={mockFiles}
        trackedFiles={['main.go', 'utils.go', 'readme.md']}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
        activeTab="all"
        onTabChange={vi.fn()}
      />
    )

    // All tracked files should be visible (flat since no directories)
    expect(screen.getByText('main.go')).toBeInTheDocument()
    expect(screen.getByText('utils.go')).toBeInTheDocument()
    expect(screen.getByText('readme.md')).toBeInTheDocument()
  })

  it('does not show tabs when trackedFiles is empty', () => {
    render(
      <FileListPanel
        files={mockFiles}
        commentsByFile={new Map()}
        selectedFile={null}
        onSelectFile={vi.fn()}
        activeTab="changed"
        onTabChange={vi.fn()}
      />
    )

    expect(screen.queryByRole('tab')).not.toBeInTheDocument()
  })

  describe('reviewed checkboxes', () => {
    it('renders checkboxes when onToggleReviewed is provided', () => {
      render(
        <FileListPanel
          files={mockFiles}
          commentsByFile={new Map()}
          selectedFile={null}
          onSelectFile={vi.fn()}
          activeTab="changed"
          onTabChange={vi.fn()}
          reviewedFiles={new Set()}
          onToggleReviewed={vi.fn()}
        />
      )

      expect(screen.getByRole('checkbox', { name: /mark main\.go as reviewed/i })).toBeInTheDocument()
      expect(screen.getByRole('checkbox', { name: /mark utils\.go as reviewed/i })).toBeInTheDocument()
    })

    it('reflects checked state from reviewedFiles', () => {
      render(
        <FileListPanel
          files={mockFiles}
          commentsByFile={new Map()}
          selectedFile={null}
          onSelectFile={vi.fn()}
          activeTab="changed"
          onTabChange={vi.fn()}
          reviewedFiles={new Set(['main.go'])}
          onToggleReviewed={vi.fn()}
        />
      )

      expect(screen.getByRole('checkbox', { name: /mark main\.go as reviewed/i })).toBeChecked()
      expect(screen.getByRole('checkbox', { name: /mark utils\.go as reviewed/i })).not.toBeChecked()
    })

    it('calls onToggleReviewed when checkbox is clicked', async () => {
      const user = userEvent.setup()
      const onToggleReviewed = vi.fn()
      render(
        <FileListPanel
          files={mockFiles}
          commentsByFile={new Map()}
          selectedFile={null}
          onSelectFile={vi.fn()}
          activeTab="changed"
          onTabChange={vi.fn()}
          reviewedFiles={new Set()}
          onToggleReviewed={onToggleReviewed}
        />
      )

      await user.click(screen.getByRole('checkbox', { name: /mark main\.go as reviewed/i }))

      expect(onToggleReviewed).toHaveBeenCalledWith('main.go')
    })

    it('does not call onSelectFile when checkbox is clicked', async () => {
      const user = userEvent.setup()
      const onSelectFile = vi.fn()
      render(
        <FileListPanel
          files={mockFiles}
          commentsByFile={new Map()}
          selectedFile={null}
          onSelectFile={onSelectFile}
          activeTab="changed"
          onTabChange={vi.fn()}
          reviewedFiles={new Set()}
          onToggleReviewed={vi.fn()}
        />
      )

      await user.click(screen.getByRole('checkbox', { name: /mark main\.go as reviewed/i }))

      expect(onSelectFile).not.toHaveBeenCalled()
    })

    it('does not render checkboxes when onToggleReviewed is not provided', () => {
      render(
        <FileListPanel
          files={mockFiles}
          commentsByFile={new Map()}
          selectedFile={null}
          onSelectFile={vi.fn()}
          activeTab="changed"
          onTabChange={vi.fn()}
        />
      )

      expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()
    })
  })
})
