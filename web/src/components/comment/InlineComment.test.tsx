import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { InlineComment } from './InlineComment'
import type { Comment } from '../../api/types'

const mockComment: Comment = {
  id: 'c1',
  filePath: 'main.go',
  line: 10,
  body: 'This needs refactoring',
  createdAt: '2026-01-01T00:00:00Z',
}

describe('InlineComment', () => {
  it('renders comment body', () => {
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

    expect(screen.getByText('This needs refactoring')).toBeInTheDocument()
  })

  it('renders line number', () => {
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

    expect(screen.getByText(/line 10/i)).toBeInTheDocument()
  })

  it('shows inline confirmation when delete is clicked', async () => {
    const user = userEvent.setup()
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /delete/i }))

    expect(screen.getByText('Are you sure?')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /confirm delete/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel delete/i })).toBeInTheDocument()
  })

  it('calls onDelete with comment ID when confirm is clicked', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    render(<InlineComment comment={mockComment} onDelete={onDelete} />)

    await user.click(screen.getByRole('button', { name: /delete/i }))
    await user.click(screen.getByRole('button', { name: /confirm delete/i }))

    expect(onDelete).toHaveBeenCalledWith('c1')
  })

  it('shows edit form when edit button is clicked', async () => {
    const user = userEvent.setup()
    const onUpdate = vi.fn()
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} onUpdate={onUpdate} />)

    await user.click(screen.getByRole('button', { name: /edit/i }))

    expect(screen.getByPlaceholderText(/comment/i)).toHaveValue('This needs refactoring')
    expect(screen.getByRole('button', { name: /update/i })).toBeInTheDocument()
  })

  it('calls onUpdate with id and new body when edit form is submitted', async () => {
    const user = userEvent.setup()
    const onUpdate = vi.fn()
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} onUpdate={onUpdate} />)

    await user.click(screen.getByRole('button', { name: /edit/i }))
    const textarea = screen.getByPlaceholderText(/comment/i)
    await user.clear(textarea)
    await user.type(textarea, 'Updated text')
    await user.click(screen.getByRole('button', { name: /update/i }))

    expect(onUpdate).toHaveBeenCalledWith('c1', 'Updated text')
  })

  it('does not show edit button when onUpdate is not provided', () => {
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

    expect(screen.queryByRole('button', { name: /edit/i })).not.toBeInTheDocument()
  })

  it('hides confirmation when cancel is clicked', async () => {
    const user = userEvent.setup()
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /delete/i }))
    await user.click(screen.getByRole('button', { name: /cancel delete/i }))

    expect(screen.queryByText('Are you sure?')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
  })

  describe('file-level comment (line=0)', () => {
    const fileComment: Comment = {
      id: 'c2',
      filePath: 'main.go',
      line: 0,
      body: 'General feedback',
      createdAt: '2026-01-01T00:00:00Z',
    }

    it('shows "File" label instead of line number', () => {
      render(<InlineComment comment={fileComment} onDelete={vi.fn()} />)

      expect(screen.getByText('File')).toBeInTheDocument()
      expect(screen.queryByText(/Line 0/)).not.toBeInTheDocument()
    })

    it('copies without line number for file-level comment', async () => {
      const user = userEvent.setup()
      const writeText = vi.fn().mockResolvedValue(undefined)
      vi.spyOn(navigator.clipboard, 'writeText').mockImplementation(writeText)

      render(<InlineComment comment={fileComment} onDelete={vi.fn()} />)

      await user.click(screen.getByRole('button', { name: /copy/i }))

      expect(writeText).toHaveBeenCalledWith('main.go\nGeneral feedback')
    })
  })

  describe('copy button', () => {
    it('renders a copy button', () => {
      render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

      expect(screen.getByRole('button', { name: /copy/i })).toBeInTheDocument()
    })

    it('copies filePath:line and body to clipboard when clicked', async () => {
      const user = userEvent.setup()
      const writeText = vi.fn().mockResolvedValue(undefined)
      vi.spyOn(navigator.clipboard, 'writeText').mockImplementation(writeText)

      render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

      await user.click(screen.getByRole('button', { name: /copy/i }))

      expect(writeText).toHaveBeenCalledWith('main.go:10\nThis needs refactoring')
    })

    it('shows "Copied!" feedback after clicking copy', async () => {
      const user = userEvent.setup()
      vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue(undefined)

      render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

      await user.click(screen.getByRole('button', { name: /copy/i }))

      expect(screen.getByText('Copied!')).toBeInTheDocument()
    })
  })
})
