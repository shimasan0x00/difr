import { useEffect, useRef, useState } from 'react'
import type { Comment, ReviewCategory, Severity } from '../../api/types'
import { CommentForm } from './CommentForm'
import { formatCommentPrefix } from '../../utils/formatComments'

interface InlineCommentProps {
  comment: Comment
  onDelete: (id: string) => void
  onUpdate?: (id: string, body: string, reviewCategory?: ReviewCategory, severity?: Severity) => void
}

const categoryColors: Record<string, string> = {
  MUST: 'text-red-400 bg-red-400/10 border border-red-400/30',
  IMO: 'text-blue-400 bg-blue-400/10 border border-blue-400/30',
  Q: 'text-yellow-400 bg-yellow-400/10 border border-yellow-400/30',
  FYI: 'text-green-400 bg-green-400/10 border border-green-400/30',
}

export function InlineComment({ comment, onDelete, onUpdate }: InlineCommentProps) {
  const [confirming, setConfirming] = useState(false)
  const [editing, setEditing] = useState(false)
  const [copied, setCopied] = useState(false)
  const copyTimerRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  useEffect(() => {
    return () => { if (copyTimerRef.current) clearTimeout(copyTimerRef.current) }
  }, [])

  const handleCopy = async () => {
    const prefix = formatCommentPrefix(comment.reviewCategory, comment.severity)
    let text: string
    if (comment.line === 0) {
      text = prefix
        ? `${comment.filePath}\n${prefix}\n- ${comment.body}`
        : `${comment.filePath}\n${comment.body}`
    } else {
      text = prefix
        ? `${comment.filePath}:${comment.line}\n${prefix}\n- ${comment.body}`
        : `${comment.filePath}:${comment.line}\n${comment.body}`
    }
    await navigator.clipboard.writeText(text)
    setCopied(true)
    if (copyTimerRef.current) clearTimeout(copyTimerRef.current)
    copyTimerRef.current = setTimeout(() => setCopied(false), 2000)
  }

  if (editing) {
    return (
      <CommentForm
        initialBody={comment.body}
        initialCategory={comment.reviewCategory}
        initialSeverity={comment.severity}
        onSubmit={(body, reviewCategory, severity) => {
          onUpdate?.(comment.id, body, reviewCategory, severity)
          setEditing(false)
        }}
        onCancel={() => setEditing(false)}
      />
    )
  }

  return (
    <div className="p-3 bg-[#161b22] border border-gray-700 rounded-md text-sm">
      <div className="flex items-center justify-between mb-1">
        <div className="flex items-center gap-1.5">
          <span className="text-gray-500 text-xs">{comment.line === 0 ? 'File' : `Line ${comment.line}`}</span>
          {comment.reviewCategory && (
            <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${categoryColors[comment.reviewCategory]}`}>
              {comment.reviewCategory}
            </span>
          )}
          {comment.severity && (
            <span className="text-gray-400 text-xs">
              {comment.severity}
            </span>
          )}
        </div>
        {confirming ? (
          <div className="flex items-center gap-2">
            <span className="text-gray-400 text-xs">Are you sure?</span>
            <button
              type="button"
              onClick={() => { onDelete(comment.id); setConfirming(false) }}
              className="text-red-400 hover:text-red-300 text-xs font-medium"
              aria-label="Confirm delete"
            >
              Confirm
            </button>
            <button
              type="button"
              onClick={() => setConfirming(false)}
              className="text-gray-400 hover:text-gray-200 text-xs"
              aria-label="Cancel delete"
            >
              Cancel
            </button>
          </div>
        ) : (
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleCopy}
              aria-label="Copy"
              className="text-gray-500 hover:text-blue-400 text-xs"
            >
              {copied ? 'Copied!' : 'Copy'}
            </button>
            {onUpdate && (
              <button
                type="button"
                onClick={() => setEditing(true)}
                aria-label="Edit"
                className="text-gray-500 hover:text-blue-400 text-xs"
              >
                Edit
              </button>
            )}
            <button
              type="button"
              onClick={() => setConfirming(true)}
              aria-label="Delete"
              className="text-gray-500 hover:text-red-400 text-xs"
            >
              Delete
            </button>
          </div>
        )}
      </div>
      <p className="text-gray-200">{comment.body}</p>
    </div>
  )
}
