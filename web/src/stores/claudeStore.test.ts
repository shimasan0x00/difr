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

    it('includes sessionId when available', async () => {
      useClaudeStore.getState().connect()
      await vi.advanceTimersByTimeAsync(10)
      useClaudeStore.setState({ sessionId: 'sess-abc' })

      useClaudeStore.getState().sendChat('follow up')

      const sent = JSON.parse(mockWsInstance!.sentMessages[0])
      expect(sent.sessionId).toBe('sess-abc')
    })

    it('sets error when WebSocket is not connected', () => {
      useClaudeStore.getState().sendChat('hello')

      expect(useClaudeStore.getState().messages).toHaveLength(0)
      expect(useClaudeStore.getState().loading).toBe(false)
      expect(useClaudeStore.getState().error).toBe('Not connected to Claude')
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
  })
})
