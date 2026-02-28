import { useState, useRef, useEffect, useMemo } from 'react'
import type { ChatMessage } from '../../api/types'

/** Minimal markdown renderer for code blocks and inline code. */
function SimpleMarkdown({ text }: { text: string }) {
  const parts = useMemo(() => {
    const result: { type: 'text' | 'codeblock' | 'code'; content: string; lang?: string }[] = []
    const codeBlockRe = /```(\w*)\n([\s\S]*?)```/g
    let lastIndex = 0
    let match: RegExpExecArray | null

    while ((match = codeBlockRe.exec(text)) !== null) {
      if (match.index > lastIndex) {
        result.push({ type: 'text', content: text.slice(lastIndex, match.index) })
      }
      result.push({ type: 'codeblock', content: match[2], lang: match[1] || undefined })
      lastIndex = match.index + match[0].length
    }
    if (lastIndex < text.length) {
      result.push({ type: 'text', content: text.slice(lastIndex) })
    }
    return result
  }, [text])

  return (
    <>
      {parts.map((part, i) =>
        part.type === 'codeblock' ? (
          <pre key={i} className="bg-gray-800 rounded p-2 my-1 text-xs overflow-x-auto">
            <code>{part.content}</code>
          </pre>
        ) : (
          <span key={i}>{part.content}</span>
        ),
      )}
    </>
  )
}

interface ChatPanelProps {
  onSend: (content: string) => void
  messages: ChatMessage[]
  loading: boolean
}

export function ChatPanel({ onSend, messages, loading }: ChatPanelProps) {
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (messagesEndRef.current && typeof messagesEndRef.current.scrollIntoView === 'function') {
      messagesEndRef.current.scrollIntoView({ behavior: loading ? 'instant' : 'smooth' })
    }
  }, [messages, loading])

  const handleSend = () => {
    if (input.trim() === '') return
    onSend(input)
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex flex-col h-full border border-gray-700 rounded-md bg-[#0d1117]">
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`text-sm ${
              msg.role === 'user'
                ? 'text-blue-300 text-right'
                : 'text-gray-200'
            }`}
          >
            <span className="text-xs text-gray-500 block mb-0.5">
              {msg.role === 'user' ? 'You' : 'Claude'}
            </span>
            <div className="whitespace-pre-wrap [&>code]:bg-gray-800 [&>code]:px-1 [&>code]:rounded [&>code]:text-xs [&>code]:font-mono">
              {msg.role === 'assistant' ? <SimpleMarkdown text={msg.content} /> : msg.content}
            </div>
          </div>
        ))}
        {loading && (
          <div className="text-gray-400 text-sm animate-pulse">
            Thinking...
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>
      <div className="border-t border-gray-700 p-3 flex gap-2">
        <textarea
          className="flex-1 bg-[#161b22] border border-gray-700 rounded px-3 py-2 text-sm text-gray-200 resize-none focus:outline-none focus:border-blue-500"
          rows={2}
          placeholder="Ask Claude..."
          aria-label="Message to Claude"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={loading}
        />
        <button
          type="button"
          onClick={handleSend}
          disabled={loading}
          className="px-4 py-2 text-sm bg-blue-700 text-white rounded hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed self-end"
        >
          Send
        </button>
      </div>
    </div>
  )
}
