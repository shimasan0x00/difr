import { create } from 'zustand'
import type { ChatMessage, WSMessage, WSResponse } from '../api/types'

interface ClaudeState {
  messages: ChatMessage[]
  sessionId: string | null
  loading: boolean
  error: string | null
  connected: boolean
  addMessage: (msg: ChatMessage) => void
  setSessionId: (id: string) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearMessages: () => void
  sendChat: (content: string) => void
  sendReview: (diffContent: string) => void
  connect: () => void
  disconnect: () => void
}

let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempt = 0
let isReconnecting = false
let messageCounter = 0
const MAX_RECONNECT_ATTEMPTS = 5

/** Reset module-level state. For testing only. */
export function _resetModuleState(): void {
  ws = null
  if (reconnectTimer) clearTimeout(reconnectTimer)
  reconnectTimer = null
  reconnectAttempt = 0
  isReconnecting = false
  messageCounter = 0
}

function nextMessageId(): string {
  return `msg-${++messageCounter}`
}

function getWsUrl(): string {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${location.host}/ws/claude`
}

function getReconnectDelay(attempt: number): number {
  return Math.min(1000 * 2 ** attempt, 30000)
}

export const useClaudeStore = create<ClaudeState>((set, get) => ({
  messages: [],
  sessionId: null,
  loading: false,
  error: null,
  connected: false,
  addMessage: (msg) => set((state) => ({ messages: [...state.messages, { ...msg, id: msg.id || nextMessageId(), sessionId: msg.sessionId ?? state.sessionId }] })),
  setSessionId: (id) => set({ sessionId: id }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error, loading: false }),
  clearMessages: () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const msg: WSMessage = { type: 'clear', content: '' }
      try {
        ws.send(JSON.stringify(msg))
      } catch {
        // Ignore send errors during clear
      }
    }
    set({ messages: [], sessionId: null, error: null, loading: false })
  },

  connect: () => {
    // Clear any pending reconnect timer to prevent stale timers from firing
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return

    // Reset counter only for user-initiated connections (not auto-reconnect)
    if (!isReconnecting) {
      reconnectAttempt = 0
    }
    isReconnecting = false

    const socket = new WebSocket(getWsUrl())
    ws = socket

    socket.onopen = () => {
      if (ws !== socket) return // stale connection
      reconnectAttempt = 0
      set({ connected: true, error: null })
    }

    socket.onmessage = (event) => {
      if (ws !== socket) return // stale connection
      let resp: WSResponse
      try {
        resp = JSON.parse(event.data)
      } catch {
        set({ error: 'Invalid message from server', loading: false })
        return
      }
      const state = get()

      switch (resp.type) {
        case 'session':
          if (resp.sessionId) set({ sessionId: resp.sessionId })
          break
        case 'text':
          if (resp.content) {
            const msgs = state.messages
            const last = msgs[msgs.length - 1]
            if (last && last.role === 'assistant') {
              // Update in-place for the current streaming message, then shallow-copy the array
              const updated = { ...last, content: last.content + resp.content }
              set({ messages: [...msgs.slice(0, -1), updated] })
            } else {
              set({ messages: [...msgs, { id: nextMessageId(), role: 'assistant', content: resp.content, sessionId: state.sessionId }] })
            }
          }
          break
        case 'done':
          set((prev) => ({
            loading: false,
            sessionId: resp.sessionId || prev.sessionId,
          }))
          break
        case 'error':
          set({ error: resp.error ?? 'Unknown error', loading: false })
          break
      }
    }

    socket.onclose = () => {
      if (ws !== socket) return // stale connection
      set({ connected: false, loading: false })
      ws = null

      // Auto-reconnect with exponential backoff
      if (reconnectAttempt < MAX_RECONNECT_ATTEMPTS) {
        const delay = getReconnectDelay(reconnectAttempt)
        reconnectAttempt++
        reconnectTimer = setTimeout(() => {
          reconnectTimer = null
          isReconnecting = true
          get().connect()
        }, delay)
      }
    }

    socket.onerror = () => {
      if (ws !== socket) return // stale connection
      set({ error: 'WebSocket connection failed', connected: false, loading: false })
    }
  },

  disconnect: () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempt = MAX_RECONNECT_ATTEMPTS // prevent auto-reconnect
    if (ws) {
      ws.close()
      ws = null
    }
    set({ connected: false, messages: [], sessionId: null, error: null })
  },

  sendChat: (content) => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      set({ error: 'Not connected to Claude' })
      return
    }

    const msgId = nextMessageId()
    set((state) => ({
      messages: [...state.messages, { id: msgId, role: 'user' as const, content, sessionId: state.sessionId }],
      loading: true,
      error: null,
    }))

    const msg: WSMessage = {
      type: 'chat',
      content,
    }
    try {
      ws.send(JSON.stringify(msg))
    } catch {
      set((state) => ({
        messages: state.messages.filter((m) => m.id !== msgId),
        error: 'Failed to send message',
        loading: false,
      }))
    }
  },

  sendReview: (diffContent) => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      set({ error: 'Not connected to Claude', loading: false })
      return
    }

    set({ loading: true, error: null })

    const msg: WSMessage = {
      type: 'review',
      content: diffContent,
    }
    try {
      ws.send(JSON.stringify(msg))
    } catch {
      set({ error: 'Failed to send message', loading: false })
    }
  },
}))
