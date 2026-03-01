import { useState } from 'react'

interface CommentFormProps {
  onSubmit: (body: string) => void
  onCancel: () => void
  saving?: boolean
  initialBody?: string
}

export function CommentForm({ onSubmit, onCancel, saving = false, initialBody }: CommentFormProps) {
  const [body, setBody] = useState(initialBody ?? '')

  const handleSubmit = () => {
    if (body.trim() === '' || saving) return
    onSubmit(body)
    setBody('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
    <div className="p-3 bg-[#161b22] border border-gray-700 rounded-md space-y-2">
      <textarea
        className="w-full bg-[#0d1117] border border-gray-700 rounded px-3 py-2 text-sm text-gray-200 resize-none focus:outline-none focus:border-blue-500"
        rows={3}
        placeholder="Add a comment..."
        aria-label="Comment body"
        value={body}
        onChange={(e) => setBody(e.target.value)}
        onKeyDown={handleKeyDown}
      />
      <div className="flex gap-2 justify-end">
        <button
          type="button"
          onClick={onCancel}
          className="px-3 py-1 text-sm text-gray-400 hover:text-gray-200"
        >
          Cancel
        </button>
        <button
          type="button"
          onClick={handleSubmit}
          disabled={saving}
          className="px-3 py-1 text-sm bg-green-700 text-white rounded hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {saving ? 'Saving...' : initialBody ? 'Update' : 'Add Comment'}
        </button>
      </div>
    </div>
  )
}
