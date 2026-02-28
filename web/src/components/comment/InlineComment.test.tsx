import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { InlineComment } from './InlineComment'
import type { Comment } from '../../api/types'

const mockComment: Comment = {
  id: 'c1',
  filePath: 'main.go',
  line: 10,
  body: 'This needs refactoring',
  createdAt: '2026-01-01T00:00:00Z',
}

let confirmSpy: ReturnType<typeof vi.spyOn>

describe('InlineComment', () => {
  beforeEach(() => {
    confirmSpy = vi.spyOn(window, 'confirm')
  })

  afterEach(() => {
    confirmSpy.mockRestore()
  })

  it('renders comment body', () => {
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

    expect(screen.getByText('This needs refactoring')).toBeInTheDocument()
  })

  it('renders line number', () => {
    render(<InlineComment comment={mockComment} onDelete={vi.fn()} />)

    expect(screen.getByText(/line 10/i)).toBeInTheDocument()
  })

  it('calls onDelete with comment ID when delete is confirmed', async () => {
    confirmSpy.mockReturnValue(true)
    const user = userEvent.setup()
    const onDelete = vi.fn()
    render(<InlineComment comment={mockComment} onDelete={onDelete} />)

    await user.click(screen.getByRole('button', { name: /delete/i }))

    expect(onDelete).toHaveBeenCalledWith('c1')
  })

  it('does not call onDelete when delete is cancelled', async () => {
    confirmSpy.mockReturnValue(false)
    const user = userEvent.setup()
    const onDelete = vi.fn()
    render(<InlineComment comment={mockComment} onDelete={onDelete} />)

    await user.click(screen.getByRole('button', { name: /delete/i }))

    expect(onDelete).not.toHaveBeenCalled()
  })
})
