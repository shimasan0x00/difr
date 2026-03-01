import { useState } from 'react'
import type { Comment } from '../../api/types'
import { CommentForm } from './CommentForm'

interface InlineCommentProps {
  comment: Comment
  onDelete: (id: string) => void
  onUpdate?: (id: string, body: string) => void
}

export function InlineComment({ comment, onDelete, onUpdate }: InlineCommentProps) {
  const [confirming, setConfirming] = useState(false)
  const [editing, setEditing] = useState(false)

  if (editing) {
    return (
      <CommentForm
        initialBody={comment.body}
        onSubmit={(body) => {
          onUpdate?.(comment.id, body)
          setEditing(false)
        }}
        onCancel={() => setEditing(false)}
      />
    )
  }

  return (
    <div className="p-3 bg-[#161b22] border border-gray-700 rounded-md text-sm">
      <div className="flex items-center justify-between mb-1">
        <span className="text-gray-500 text-xs">Line {comment.line}</span>
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
