import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { useClaudeStore, _resetModuleState } from './claudeStore'

// --- WebSocket mock ---
type WSEventHandler = ((event: { data: string }) => void) | (() => void) | null

class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  readyState = MockWebSocket.CONNECTING
  onopen: (() => void) | null = null
  onclose: (() => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null
  onerror: (() => void) | null = null
  sentMessages: string[] = []

  constructor(_url: string) {
    // Schedule async open to mimic real WebSocket
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN
      this.onopen?.()
    }, 0)
  }

  send(data: string) {
    this.sentMessages.push(data)
  }

  close() {
    this.readyState = MockWebSocket.CLOSED
    this.onclose?.()
  }

  // Test helpers
  simulateMessage(data: unknown) {
    this.onmessage?.({ data: JSON.stringify(data) })
  }

  simulateError() {
    this.onerror?.()
  }

  simulateClose() {
    this.readyState = MockWebSocket.CLOSED
    this.onclose?.()
  }
}

let mockWsInstance: MockWebSocket | null = null

function resetStore() {
  _resetModuleState()
  useClaudeStore.setState({
    messages: [],
    sessionId: null,
    loading: false,
    error: null,
    connected: false,
  })
}

describe('claudeStore', () => {
  beforeEach(() => {
    resetStore()
    mockWsInstance = null
    vi.useFakeTimers()

    // Mock WebSocket constructor
    vi.stubGlobal('WebSocket', class extends MockWebSocket {
      constructor(url: string) {
        super(url)
        mockWsInstance = this
      }
    })
    // Needed for static properties
    ;(globalThis as Record<string, unknown>).WebSocket = Object.assign(
      (globalThis as Record<string, unknown>).WebSocket as object,
      { OPEN: 1, CONNECTING: 0, CLOSING: 2, CLOSED: 3 },
    )
  })

  afterEach(() => {
    useClaudeStore.getState().disconnect()
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  describe('connect', () => {
    it('creates WebSocket and sets connected on open', async () => {
      useClaudeStore.getState().connect()

      expect(mockWsInstance).not.toBeNull()

      // Trigger onopen
      await vi.advanceTimersByTimeAsync(10)

      expect(useClaudeStore.getState().connected).toBe(true)
      expect(useClaudeStore.getState().error).toBeNull()
    })

    it('does not create duplicate connection when already OPEN', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      const firstInstance = mockWsInstance
      useClaudeStore.getState().connect()

      expect(mockWsInstance).toBe(firstInstance)
    })

    it('does not create duplicate connection when CONNECTING', () => {
      useClaudeStore.getState().connect()
      // Don't advance timers - still CONNECTING

      const firstInstance = mockWsInstance
      useClaudeStore.getState().connect()

      expect(mockWsInstance).toBe(firstInstance)
    })
  })

  describe('disconnect', () => {
    it('closes WebSocket, clears messages, and sets connected to false', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({
        messages: [{ id: 'test-1', role: 'user', content: 'hi' }],
        sessionId: 'sess-1',
      })

      useClaudeStore.getState().disconnect()

      const state = useClaudeStore.getState()
      expect(state.connected).toBe(false)
      expect(state.messages).toHaveLength(0)
      expect(state.sessionId).toBeNull()
    })

    it('prevents auto-reconnect after disconnect', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      useClaudeStore.getState().disconnect()

      // Even after waiting, should not reconnect
      await vi.advanceTimersByTimeAsync(60000)
      expect(useClaudeStore.getState().connected).toBe(false)
    })
  })

  describe('onmessage', () => {
    it('handles session message and stores sessionId', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.simulateMessage({ type: 'session', sessionId: 'sess-123' })

      expect(useClaudeStore.getState().sessionId).toBe('sess-123')
    })

    it('handles text message by appending to new assistant message', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.simulateMessage({ type: 'text', content: 'Hello' })

      const messages = useClaudeStore.getState().messages
      expect(messages).toHaveLength(1)
      expect(messages[0].role).toBe('assistant')
      expect(messages[0].content).toBe('Hello')
      expect(messages[0].id).toBeTruthy()
    })

    it('handles text message by appending to existing assistant message', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.simulateMessage({ type: 'text', content: 'Hello ' })
      mockWsInstance!.simulateMessage({ type: 'text', content: 'world' })

      const messages = useClaudeStore.getState().messages
      expect(messages).toHaveLength(1)
      expect(messages[0].content).toBe('Hello world')
    })

    it('handles done message and clears loading', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({ loading: true })

      mockWsInstance!.simulateMessage({ type: 'done' })

      expect(useClaudeStore.getState().loading).toBe(false)
    })

    it('handles done message with sessionId', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({ loading: true })

      mockWsInstance!.simulateMessage({ type: 'done', sessionId: 'sess-from-done' })

      expect(useClaudeStore.getState().loading).toBe(false)
      expect(useClaudeStore.getState().sessionId).toBe('sess-from-done')
    })

    it('preserves existing sessionId when done has no sessionId', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({ loading: true, sessionId: 'existing-sess' })

      mockWsInstance!.simulateMessage({ type: 'done' })

      expect(useClaudeStore.getState().sessionId).toBe('existing-sess')
    })

    it('handles error message and sets error state', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.simulateMessage({ type: 'error', error: 'Claude CLI not available' })

      expect(useClaudeStore.getState().error).toBe('Claude CLI not available')
      expect(useClaudeStore.getState().loading).toBe(false)
    })

    it('handles invalid JSON by setting error', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      // Directly call onmessage with invalid JSON
      mockWsInstance!.onmessage?.({ data: 'not-json' })

      expect(useClaudeStore.getState().error).toBe('Invalid message from server')
    })
  })

  describe('onclose / reconnect', () => {
    it('sets connected to false on close', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.simulateClose()

      expect(useClaudeStore.getState().connected).toBe(false)
    })

    it('resets loading to false on close', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({ loading: true })

      mockWsInstance!.simulateClose()

      expect(useClaudeStore.getState().loading).toBe(false)
    })

    it('stops reconnecting after MAX_RECONNECT_ATTEMPTS (5) reached', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10) // onopen fires

      let wsCreationCount = 1

      // Replace with a minimal mock that does NOT schedule onopen (simulates connection failure)
      class FailingWebSocket {
        static CONNECTING = 0
        static OPEN = 1
        static CLOSING = 2
        static CLOSED = 3
        readyState = 0
        onopen: (() => void) | null = null
        onclose: (() => void) | null = null
        onmessage: ((event: { data: string }) => void) | null = null
        onerror: (() => void) | null = null
        constructor(_url: string) {
          mockWsInstance = this as unknown as MockWebSocket
          wsCreationCount++
        }
        send() { /* noop */ }
        close() {
          this.readyState = FailingWebSocket.CLOSED
          this.onclose?.()
        }
        simulateClose() { this.close() }
      }
      vi.stubGlobal('WebSocket', FailingWebSocket)

      // Trigger close → reconnect cycle
      for (let i = 0; i < 5; i++) {
        mockWsInstance!.simulateClose()
        const delay = Math.min(1000 * 2 ** i, 30000)
        await vi.advanceTimersByTimeAsync(delay + 10)
      }

      expect(wsCreationCount).toBe(6) // 1 initial + 5 reconnects

      // 6th close: should NOT trigger reconnect (reconnectAttempt >= MAX)
      const countBefore = wsCreationCount
      mockWsInstance!.simulateClose()
      await vi.advanceTimersByTimeAsync(60000)

      expect(wsCreationCount).toBe(countBefore)
      expect(useClaudeStore.getState().connected).toBe(false)
    })

    it('clears pending reconnect timer when connect is called manually', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      // Close triggers a reconnect timer (1000ms delay)
      mockWsInstance!.simulateClose()
      const closedInstance = mockWsInstance

      // Before timer fires, manually connect
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10) // onopen fires for new connection
      const manualInstance = mockWsInstance
      expect(manualInstance).not.toBe(closedInstance)

      // Advance past the old reconnect timer — should NOT create another connection
      await vi.advanceTimersByTimeAsync(2000)
      expect(mockWsInstance).toBe(manualInstance)
    })

    it('auto-reconnects with exponential backoff', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.simulateClose()
      const firstInstance = mockWsInstance

      // First reconnect delay: 1000ms * 2^0 = 1000ms
      await vi.advanceTimersByTimeAsync(1000)

      expect(mockWsInstance).not.toBe(firstInstance)
    })
  })

  describe('onerror', () => {
    it('sets error state on WebSocket error', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.simulateError()

      expect(useClaudeStore.getState().error).toBe('WebSocket connection failed')
      expect(useClaudeStore.getState().connected).toBe(false)
    })
  })

  describe('sendChat', () => {
    it('adds user message, sets loading, and sends via WebSocket', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      useClaudeStore.getState().sendChat('Hello Claude')

      const state = useClaudeStore.getState()
      expect(state.messages).toHaveLength(1)
      expect(state.messages[0].role).toBe('user')
      expect(state.messages[0].content).toBe('Hello Claude')
      expect(state.messages[0].id).toBeTruthy()
      expect(state.loading).toBe(true)

      const sent = JSON.parse(mockWsInstance!.sentMessages[0])
      expect(sent.type).toBe('chat')
      expect(sent.content).toBe('Hello Claude')
    })

    it('does not include sessionId in sent message', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({ sessionId: 'sess-abc' })

      useClaudeStore.getState().sendChat('follow up')

      const sent = JSON.parse(mockWsInstance!.sentMessages[0])
      expect(sent.sessionId).toBeUndefined()
      expect(sent.type).toBe('chat')
      expect(sent.content).toBe('follow up')
    })

    it('sets error when WebSocket is not connected', () => {
      useClaudeStore.getState().sendChat('hello')

      expect(useClaudeStore.getState().messages).toHaveLength(0)
      expect(useClaudeStore.getState().loading).toBe(false)
      expect(useClaudeStore.getState().error).toBe('Not connected to Claude')
    })

    it('sets error and removes user message when ws.send() throws', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      // Make send throw
      mockWsInstance!.send = () => { throw new Error('send failed') }

      useClaudeStore.getState().sendChat('hello')

      const state = useClaudeStore.getState()
      expect(state.error).toBe('Failed to send message')
      expect(state.loading).toBe(false)
      expect(state.messages).toHaveLength(0)
    })

    it('only removes the failed message and keeps existing messages on send failure', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      // Send a successful message first
      useClaudeStore.getState().sendChat('first message')
      expect(useClaudeStore.getState().messages).toHaveLength(1)
      expect(useClaudeStore.getState().messages[0].content).toBe('first message')

      // Make send throw for the second message
      mockWsInstance!.send = () => { throw new Error('send failed') }

      useClaudeStore.getState().sendChat('failed message')

      const state = useClaudeStore.getState()
      expect(state.messages).toHaveLength(1)
      expect(state.messages[0].content).toBe('first message')
      expect(state.error).toBe('Failed to send message')
      expect(state.loading).toBe(false)
    })
  })

  describe('sendReview', () => {
    it('sets loading and sends review message via WebSocket', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      useClaudeStore.getState().sendReview('diff content')

      expect(useClaudeStore.getState().loading).toBe(true)

      const sent = JSON.parse(mockWsInstance!.sentMessages[0])
      expect(sent.type).toBe('review')
      expect(sent.content).toBe('diff content')
    })

    it('sets error when WebSocket is not connected', () => {
      useClaudeStore.getState().sendReview('diff content')

      expect(useClaudeStore.getState().loading).toBe(false)
      expect(useClaudeStore.getState().error).toBe('Not connected to Claude')
    })

    it('sets error when ws.send() throws', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.send = () => { throw new Error('send failed') }

      useClaudeStore.getState().sendReview('diff content')

      const state = useClaudeStore.getState()
      expect(state.error).toBe('Failed to send message')
      expect(state.loading).toBe(false)
    })
  })

  describe('sessionId on messages', () => {
    it('sendChat attaches current sessionId to user message', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({ sessionId: 'sess-A' })

      useClaudeStore.getState().sendChat('hello')

      expect(useClaudeStore.getState().messages[0].sessionId).toBe('sess-A')
    })

    it('text response creates assistant message with current sessionId', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)

      mockWsInstance!.simulateMessage({ type: 'session', sessionId: 'sess-A' })
      mockWsInstance!.simulateMessage({ type: 'text', content: 'Hi' })

      expect(useClaudeStore.getState().messages[0].sessionId).toBe('sess-A')
    })

    it('streaming text preserves original sessionId on append', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({ sessionId: 'sess-A' })

      mockWsInstance!.simulateMessage({ type: 'text', content: 'Hello ' })
      mockWsInstance!.simulateMessage({ type: 'text', content: 'world' })

      const msgs = useClaudeStore.getState().messages
      expect(msgs).toHaveLength(1)
      expect(msgs[0].sessionId).toBe('sess-A')
    })

    it('addMessage defaults to current sessionId when not provided', () => {
      useClaudeStore.setState({ sessionId: 'sess-B' })

      useClaudeStore.getState().addMessage({ id: 'x', role: 'user', content: 'test' })

      expect(useClaudeStore.getState().messages[0].sessionId).toBe('sess-B')
    })

    it('addMessage preserves explicit sessionId when provided', () => {
      useClaudeStore.setState({ sessionId: 'sess-B' })

      useClaudeStore.getState().addMessage({ id: 'x', role: 'user', content: 'test', sessionId: 'sess-A' })

      expect(useClaudeStore.getState().messages[0].sessionId).toBe('sess-A')
    })
  })

  describe('state mutations', () => {
    it('addMessage appends to messages array', () => {
      useClaudeStore.getState().addMessage({ id: 'test-1', role: 'user', content: 'hi' })
      useClaudeStore.getState().addMessage({ id: 'test-2', role: 'assistant', content: 'hello' })

      expect(useClaudeStore.getState().messages).toHaveLength(2)
    })

    it('clearMessages resets messages, sessionId, and error', () => {
      useClaudeStore.setState({
        messages: [{ id: 'test-1', role: 'user', content: 'hi' }],
        sessionId: 'sess-1',
        error: 'some error',
      })

      useClaudeStore.getState().clearMessages()

      const state = useClaudeStore.getState()
      expect(state.messages).toHaveLength(0)
      expect(state.sessionId).toBeNull()
      expect(state.error).toBeNull()
    })

    it('clearMessages sends clear message via WebSocket when connected', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({
        messages: [{ id: 'test-1', role: 'user', content: 'hi' }],
        sessionId: 'sess-1',
      })

      useClaudeStore.getState().clearMessages()

      expect(mockWsInstance!.sentMessages).toHaveLength(1)
      const sent = JSON.parse(mockWsInstance!.sentMessages[0])
      expect(sent.type).toBe('clear')
      expect(sent.content).toBe('')

      const state = useClaudeStore.getState()
      expect(state.messages).toHaveLength(0)
      expect(state.sessionId).toBeNull()
    })

    it('clearMessages does not send WebSocket message when not connected', () => {
      useClaudeStore.setState({
        messages: [{ id: 'test-1', role: 'user', content: 'hi' }],
      })

      useClaudeStore.getState().clearMessages()

      // No WebSocket instance, so no messages sent
      expect(useClaudeStore.getState().messages).toHaveLength(0)
    })
  })
})
