import { useState } from 'react'
import type { ReviewCategory, Severity } from '../../api/types'

interface CommentFormProps {
  onSubmit: (body: string, reviewCategory?: ReviewCategory, severity?: Severity) => void
  onCancel: () => void
  saving?: boolean
  initialBody?: string
  initialCategory?: ReviewCategory
  initialSeverity?: Severity
}

export function CommentForm({ onSubmit, onCancel, saving = false, initialBody, initialCategory, initialSeverity }: CommentFormProps) {
  const [body, setBody] = useState(initialBody ?? '')
  const [category, setCategory] = useState<ReviewCategory | ''>(initialCategory ?? '')
  const [severity, setSeverity] = useState<Severity | ''>(initialSeverity ?? '')

  const handleSubmit = () => {
    if (body.trim() === '' || saving) return
    onSubmit(body, category || undefined, severity || undefined)
    setBody('')
    setCategory('')
    setSeverity('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
    <div className="p-3 bg-[#161b22] border border-gray-700 rounded-md space-y-2">
      <div className="flex gap-2">
        <select
          value={category}
          onChange={(e) => setCategory(e.target.value as ReviewCategory | '')}
          aria-label="Review category"
          className="bg-[#0d1117] border border-gray-700 rounded px-2 py-1 text-xs text-gray-200 focus:outline-none focus:border-blue-500"
        >
          <option value="">(None)</option>
          <option value="MUST">MUST</option>
          <option value="IMO">IMO</option>
          <option value="Q">Q</option>
          <option value="FYI">FYI</option>
        </select>
        <select
          value={severity}
          onChange={(e) => setSeverity(e.target.value as Severity | '')}
          aria-label="Severity"
          className="bg-[#0d1117] border border-gray-700 rounded px-2 py-1 text-xs text-gray-200 focus:outline-none focus:border-blue-500"
        >
          <option value="">(None)</option>
          <option value="Critical">Critical</option>
          <option value="High">High</option>
          <option value="Middle">Middle</option>
          <option value="Low">Low</option>
        </select>
      </div>
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
