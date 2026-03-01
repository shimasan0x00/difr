import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { CommentForm } from './CommentForm'

describe('CommentForm', () => {
  it('renders textarea and submit button', () => {
    render(<CommentForm onSubmit={vi.fn()} onCancel={vi.fn()} />)

    expect(screen.getByPlaceholderText(/comment/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /submit|add/i })).toBeInTheDocument()
  })

  it('calls onSubmit with body text when submitted', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    render(<CommentForm onSubmit={onSubmit} onCancel={vi.fn()} />)

    await user.type(screen.getByPlaceholderText(/comment/i), 'This needs refactoring')
    await user.click(screen.getByRole('button', { name: /submit|add/i }))

    expect(onSubmit).toHaveBeenCalledWith('This needs refactoring')

    // Verify textarea is cleared after submit
    expect(screen.getByPlaceholderText(/comment/i)).toHaveValue('')
  })

  it('does not submit when body is empty', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    render(<CommentForm onSubmit={onSubmit} onCancel={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /submit|add/i }))

    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('does not submit when body contains only whitespace', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    render(<CommentForm onSubmit={onSubmit} onCancel={vi.fn()} />)

    await user.type(screen.getByPlaceholderText(/comment/i), '   \t  ')
    await user.click(screen.getByRole('button', { name: /submit|add/i }))

    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('calls onCancel when cancel button is clicked', async () => {
    const user = userEvent.setup()
    const onCancel = vi.fn()
    render(<CommentForm onSubmit={vi.fn()} onCancel={onCancel} />)

    await user.click(screen.getByRole('button', { name: /cancel/i }))

    expect(onCancel).toHaveBeenCalled()
  })

  it('pre-fills textarea with initialBody and shows Update button', () => {
    render(<CommentForm onSubmit={vi.fn()} onCancel={vi.fn()} initialBody="Existing comment" />)

    expect(screen.getByPlaceholderText(/comment/i)).toHaveValue('Existing comment')
    expect(screen.getByRole('button', { name: /update/i })).toBeInTheDocument()
  })

  it('submits updated body when initialBody is provided', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    render(<CommentForm onSubmit={onSubmit} onCancel={vi.fn()} initialBody="Old text" />)

    const textarea = screen.getByPlaceholderText(/comment/i)
    await user.clear(textarea)
    await user.type(textarea, 'New text')
    await user.click(screen.getByRole('button', { name: /update/i }))

    expect(onSubmit).toHaveBeenCalledWith('New text')
  })

  it('disables submit button and shows "Saving..." when saving', () => {
    render(<CommentForm onSubmit={vi.fn()} onCancel={vi.fn()} saving={true} />)

    const button = screen.getByRole('button', { name: /saving/i })
    expect(button).toBeDisabled()
    expect(button).toHaveTextContent('Saving...')
  })
})
