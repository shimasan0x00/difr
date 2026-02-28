import type { Comment } from '../../api/types'

interface InlineCommentProps {
  comment: Comment
  onDelete: (id: string) => void
}

export function InlineComment({ comment, onDelete }: InlineCommentProps) {
  const handleDelete = () => {
    if (window.confirm('Delete this comment?')) {
      onDelete(comment.id)
    }
  }

  return (
    <div className="p-3 bg-[#161b22] border border-gray-700 rounded-md text-sm">
      <div className="flex items-center justify-between mb-1">
        <span className="text-gray-500 text-xs">Line {comment.line}</span>
        <button
          type="button"
          onClick={handleDelete}
          aria-label="Delete"
          className="text-gray-500 hover:text-red-400 text-xs"
        >
          Delete
        </button>
      </div>
      <p className="text-gray-200">{comment.body}</p>
    </div>
  )
}
