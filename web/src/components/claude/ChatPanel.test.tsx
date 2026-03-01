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
    // 'Claude' appears in both header and message role label
    expect(screen.getAllByText('Claude').length).toBeGreaterThanOrEqual(2)
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

  describe('SimpleMarkdown inline code', () => {
    it('renders inline code with <code> tag', () => {
      const messages = [
        { id: 'msg-1', role: 'assistant' as const, content: 'Use `fmt.Println` here' },
      ]
      render(<ChatPanel onSend={vi.fn()} messages={messages} loading={false} />)

      const codeEl = screen.getByText('fmt.Println')
      expect(codeEl.tagName).toBe('CODE')
    })

    it('renders mixed fenced code blocks and inline code', () => {
      const messages = [
        {
          id: 'msg-1',
          role: 'assistant' as const,
          content: 'Use `foo` then:\n```go\nbar()\n```\nand `baz`',
        },
      ]
      render(<ChatPanel onSend={vi.fn()} messages={messages} loading={false} />)

      expect(screen.getByText('foo').tagName).toBe('CODE')
      expect(screen.getByText('baz').tagName).toBe('CODE')
      // fenced code block content
      expect(screen.getByText('bar()')).toBeInTheDocument()
    })

    it('does not parse backticks inside fenced code blocks as inline code', () => {
      const messages = [
        {
          id: 'msg-1',
          role: 'assistant' as const,
          content: '```\n`not inline`\n```',
        },
      ]
      render(<ChatPanel onSend={vi.fn()} messages={messages} loading={false} />)

      const el = screen.getByText('`not inline`')
      // Should be inside a <pre><code> not a standalone <code>
      expect(el.closest('pre')).not.toBeNull()
    })

    it('renders multiple inline code spans', () => {
      const messages = [
        {
          id: 'msg-1',
          role: 'assistant' as const,
          content: 'Compare `alpha` and `beta` values',
        },
      ]
      render(<ChatPanel onSend={vi.fn()} messages={messages} loading={false} />)

      expect(screen.getByText('alpha').tagName).toBe('CODE')
      expect(screen.getByText('beta').tagName).toBe('CODE')
    })
  })

  describe('ChatPanel header', () => {
    it('renders header with Claude title', () => {
      render(<ChatPanel onSend={vi.fn()} messages={[]} loading={false} />)

      expect(screen.getByText('Claude')).toBeInTheDocument()
    })

    it('shows green indicator when connected', () => {
      const { container } = render(
        <ChatPanel onSend={vi.fn()} messages={[]} loading={false} connected={true} />,
      )

      const indicator = container.querySelector('[data-testid="connection-indicator"]')
      expect(indicator).toBeInTheDocument()
      expect(indicator?.className).toContain('bg-green')
    })

    it('shows red indicator when disconnected', () => {
      const { container } = render(
        <ChatPanel onSend={vi.fn()} messages={[]} loading={false} connected={false} />,
      )

      const indicator = container.querySelector('[data-testid="connection-indicator"]')
      expect(indicator).toBeInTheDocument()
      expect(indicator?.className).toContain('bg-red')
    })

    it('shows Clear button when messages exist and onClear is provided', () => {
      const messages = [
        { id: 'msg-1', role: 'user' as const, content: 'Hello' },
      ]
      const onClear = vi.fn()
      render(
        <ChatPanel onSend={vi.fn()} messages={messages} loading={false} onClear={onClear} />,
      )

      expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument()
    })

    it('calls onClear when Clear button is clicked', async () => {
      const user = userEvent.setup()
      const messages = [
        { id: 'msg-1', role: 'user' as const, content: 'Hello' },
      ]
      const onClear = vi.fn()
      render(
        <ChatPanel onSend={vi.fn()} messages={messages} loading={false} onClear={onClear} />,
      )

      await user.click(screen.getByRole('button', { name: /clear/i }))

      expect(onClear).toHaveBeenCalled()
    })
  })
})
