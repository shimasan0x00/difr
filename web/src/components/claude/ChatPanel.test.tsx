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

    it('renders fenced code block without newline after opening backticks', () => {
      const messages = [
        {
          id: 'msg-1',
          role: 'assistant' as const,
          content: '```go func main(){}```',
        },
      ]
      render(<ChatPanel onSend={vi.fn()} messages={messages} loading={false} />)

      // lang="go", content=" func main(){}" — space separates since \n is optional
      const el = screen.getByText(/func main/)
      expect(el.closest('pre')).not.toBeNull()
    })
  })

  describe('error display', () => {
    it('shows error banner when error prop is provided', () => {
      render(
        <ChatPanel onSend={vi.fn()} messages={[]} loading={false} error="Claude CLI not available" />,
      )

      const alert = screen.getByRole('alert')
      expect(alert).toBeInTheDocument()
      expect(alert).toHaveTextContent('Claude CLI not available')
    })

    it('does not show error banner when error is null', () => {
      render(
        <ChatPanel onSend={vi.fn()} messages={[]} loading={false} error={null} />,
      )

      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })

    it('does not show error banner when error is not provided', () => {
      render(<ChatPanel onSend={vi.fn()} messages={[]} loading={false} />)

      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })
  })

  describe('keyboard shortcuts', () => {
    it('sends on Ctrl+Enter', async () => {
      const user = userEvent.setup()
      const onSend = vi.fn()
      render(<ChatPanel onSend={onSend} messages={[]} loading={false} />)

      const textarea = screen.getByPlaceholderText(/ask claude/i)
      await user.type(textarea, 'hello')
      await user.keyboard('{Control>}{Enter}{/Control}')

      expect(onSend).toHaveBeenCalledWith('hello')
    })

    it('sends on Cmd+Enter (Meta)', async () => {
      const user = userEvent.setup()
      const onSend = vi.fn()
      render(<ChatPanel onSend={onSend} messages={[]} loading={false} />)

      const textarea = screen.getByPlaceholderText(/ask claude/i)
      await user.type(textarea, 'hello')
      await user.keyboard('{Meta>}{Enter}{/Meta}')

      expect(onSend).toHaveBeenCalledWith('hello')
    })

    it('does not send on plain Enter', async () => {
      const user = userEvent.setup()
      const onSend = vi.fn()
      render(<ChatPanel onSend={onSend} messages={[]} loading={false} />)

      const textarea = screen.getByPlaceholderText(/ask claude/i)
      await user.type(textarea, 'hello')
      await user.keyboard('{Enter}')

      expect(onSend).not.toHaveBeenCalled()
    })
  })

  describe('disconnected state', () => {
    it('disables textarea and send button when disconnected', () => {
      render(<ChatPanel onSend={vi.fn()} messages={[]} loading={false} connected={false} />)

      expect(screen.getByPlaceholderText(/ask claude/i)).toBeDisabled()
      expect(screen.getByRole('button', { name: /send/i })).toBeDisabled()
    })

    it('shows error and messages when disconnected', () => {
      const messages = [
        { id: 'msg-1', role: 'user' as const, content: 'Hello' },
      ]
      render(
        <ChatPanel
          onSend={vi.fn()}
          messages={messages}
          loading={false}
          connected={false}
          error="WebSocket connection failed"
        />,
      )

      expect(screen.getByRole('alert')).toHaveTextContent('WebSocket connection failed')
      expect(screen.getByText('Hello')).toBeInTheDocument()
    })
  })

  describe('collapse button', () => {
    it('renders collapse button when onCollapse is provided', () => {
      render(<ChatPanel onSend={vi.fn()} messages={[]} loading={false} onCollapse={vi.fn()} />)

      expect(screen.getByRole('button', { name: /collapse chat panel/i })).toBeInTheDocument()
    })

    it('does not render collapse button when onCollapse is not provided', () => {
      render(<ChatPanel onSend={vi.fn()} messages={[]} loading={false} />)

      expect(screen.queryByRole('button', { name: /collapse chat panel/i })).not.toBeInTheDocument()
    })

    it('calls onCollapse when collapse button is clicked', async () => {
      const user = userEvent.setup()
      const onCollapse = vi.fn()
      render(<ChatPanel onSend={vi.fn()} messages={[]} loading={false} onCollapse={onCollapse} />)

      await user.click(screen.getByRole('button', { name: /collapse chat panel/i }))

      expect(onCollapse).toHaveBeenCalled()
    })
  })

  describe('ChatPanel header', () => {
    it('renders header with Claude title', () => {
      render(<ChatPanel onSend={vi.fn()} messages={[]} loading={false} />)

      expect(screen.getByText('Claude')).toBeInTheDocument()
    })

    it('shows green indicator when connected with active session', () => {
      const { container } = render(
        <ChatPanel onSend={vi.fn()} messages={[]} loading={false} connected={true} sessionId="abc" />,
      )

      const indicator = container.querySelector('[data-testid="connection-indicator"]')
      expect(indicator).toBeInTheDocument()
      expect(indicator?.className).toContain('bg-green')
      expect(indicator).toHaveAttribute('title', 'Session active')
    })

    it('shows yellow indicator when connected without session', () => {
      const { container } = render(
        <ChatPanel onSend={vi.fn()} messages={[]} loading={false} connected={true} sessionId={null} />,
      )

      const indicator = container.querySelector('[data-testid="connection-indicator"]')
      expect(indicator).toBeInTheDocument()
      expect(indicator?.className).toContain('bg-yellow')
      expect(indicator).toHaveAttribute('title', 'No active session')
    })

    it('shows yellow indicator when connected and sessionId not provided', () => {
      const { container } = render(
        <ChatPanel onSend={vi.fn()} messages={[]} loading={false} connected={true} />,
      )

      const indicator = container.querySelector('[data-testid="connection-indicator"]')
      expect(indicator).toBeInTheDocument()
      expect(indicator?.className).toContain('bg-yellow')
      expect(indicator).toHaveAttribute('title', 'No active session')
    })

    it('shows red indicator when disconnected', () => {
      const { container } = render(
        <ChatPanel onSend={vi.fn()} messages={[]} loading={false} connected={false} />,
      )

      const indicator = container.querySelector('[data-testid="connection-indicator"]')
      expect(indicator).toBeInTheDocument()
      expect(indicator?.className).toContain('bg-red')
      expect(indicator).toHaveAttribute('title', 'Disconnected')
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

  describe('stale session messages', () => {
    it('applies opacity-50 to messages from a previous session', () => {
      const messages = [
        { id: 'msg-1', role: 'user' as const, content: 'Old question', sessionId: 'sess-old' },
        { id: 'msg-2', role: 'assistant' as const, content: 'Old answer', sessionId: 'sess-old' },
        { id: 'msg-3', role: 'user' as const, content: 'New question', sessionId: 'sess-new' },
      ]
      render(
        <ChatPanel
          onSend={vi.fn()}
          messages={messages}
          loading={false}
          connected={true}
          sessionId="sess-new"
        />,
      )

      const staleMessages = screen.getAllByTestId('stale-message')
      expect(staleMessages).toHaveLength(2)
      expect(staleMessages[0]).toHaveTextContent('Old question')
      expect(staleMessages[1]).toHaveTextContent('Old answer')
    })

    it('does not apply opacity-50 to messages from the current session', () => {
      const messages = [
        { id: 'msg-1', role: 'user' as const, content: 'Current msg', sessionId: 'sess-current' },
      ]
      render(
        <ChatPanel
          onSend={vi.fn()}
          messages={messages}
          loading={false}
          connected={true}
          sessionId="sess-current"
        />,
      )

      expect(screen.queryByTestId('stale-message')).not.toBeInTheDocument()
    })

    it('does not grey out messages that have no sessionId', () => {
      const messages = [
        { id: 'msg-1', role: 'user' as const, content: 'Pre-session msg' },
      ]
      render(
        <ChatPanel
          onSend={vi.fn()}
          messages={messages}
          loading={false}
          connected={true}
          sessionId="sess-new"
        />,
      )

      expect(screen.queryByTestId('stale-message')).not.toBeInTheDocument()
    })

    it('does not grey out any messages when sessionId is null', () => {
      const messages = [
        { id: 'msg-1', role: 'user' as const, content: 'Some msg', sessionId: 'sess-old' },
      ]
      render(
        <ChatPanel
          onSend={vi.fn()}
          messages={messages}
          loading={false}
          connected={true}
          sessionId={null}
        />,
      )

      expect(screen.queryByTestId('stale-message')).not.toBeInTheDocument()
    })
  })
})
