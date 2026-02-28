import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { ChatPanel } from './ChatPanel'

describe('ChatPanel', () => {
  it('renders textarea and send button', () => {
    render(<ChatPanel onSend={vi.fn()} messages={[]} loading={false} />)

    expect(screen.getByPlaceholderText(/ask claude/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /send/i })).toBeInTheDocument()
  })

  it('sends input text and clears the field', async () => {
    const user = userEvent.setup()
    const onSend = vi.fn()
    render(<ChatPanel onSend={onSend} messages={[]} loading={false} />)

    const textarea = screen.getByPlaceholderText(/ask claude/i)
    await user.type(textarea, 'Review this code')
    await user.click(screen.getByRole('button', { name: /send/i }))

    expect(onSend).toHaveBeenCalledWith('Review this code')
    expect(textarea).toHaveValue('')
  })

  it('does not send when input is empty', async () => {
    const user = userEvent.setup()
    const onSend = vi.fn()
    render(<ChatPanel onSend={onSend} messages={[]} loading={false} />)

    await user.click(screen.getByRole('button', { name: /send/i }))

    expect(onSend).not.toHaveBeenCalled()
  })

  it('renders messages with correct role labels', () => {
    const messages = [
      { id: 'msg-1', role: 'user' as const, content: 'Hello' },
      { id: 'msg-2', role: 'assistant' as const, content: 'Hi there!' },
    ]
    render(<ChatPanel onSend={vi.fn()} messages={messages} loading={false} />)

    expect(screen.getByText('Hello')).toBeInTheDocument()
    expect(screen.getByText('Hi there!')).toBeInTheDocument()
    expect(screen.getByText('You')).toBeInTheDocument()
    expect(screen.getByText('Claude')).toBeInTheDocument()
  })

  it('shows thinking indicator when loading', () => {
    render(<ChatPanel onSend={vi.fn()} messages={[]} loading={true} />)

    expect(screen.getByText(/thinking/i)).toBeInTheDocument()
  })

  it('disables textarea and send button when loading', () => {
    render(<ChatPanel onSend={vi.fn()} messages={[]} loading={true} />)

    expect(screen.getByPlaceholderText(/ask claude/i)).toBeDisabled()
    expect(screen.getByRole('button', { name: /send/i })).toBeDisabled()
  })
})
