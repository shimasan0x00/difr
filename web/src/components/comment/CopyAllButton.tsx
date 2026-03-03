import { useEffect, useRef, useState } from 'react'
import type { Comment } from '../../api/types'
import { formatAllComments } from '../../utils/formatComments'

interface CopyAllButtonProps {
  comments: Comment[]
}

export function CopyAllButton({ comments }: CopyAllButtonProps) {
  const [copied, setCopied] = useState(false)
  const copyTimerRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  useEffect(() => {
    return () => { if (copyTimerRef.current) clearTimeout(copyTimerRef.current) }
  }, [])

  const handleCopy = async () => {
    const text = formatAllComments(comments)
    await navigator.clipboard.writeText(text)
    setCopied(true)
    if (copyTimerRef.current) clearTimeout(copyTimerRef.current)
    copyTimerRef.current = setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      type="button"
      onClick={handleCopy}
      className="px-3 py-1 text-xs text-gray-400 hover:text-gray-200 border border-gray-700 rounded"
      aria-label="Copy all comments"
    >
      {copied ? 'Copied!' : 'Copy All'}
    </button>
  )
}
