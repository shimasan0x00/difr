import { useState, useRef, useEffect } from 'react'

interface ClearButtonProps {
  onClearComments: () => void
  onClearReviewedFiles: () => void
  hasComments: boolean
  hasReviewedFiles: boolean
}

export function ClearButton({ onClearComments, onClearReviewedFiles, hasComments, hasReviewedFiles }: ClearButtonProps) {
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handleMouseDown = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleMouseDown)
    return () => document.removeEventListener('mousedown', handleMouseDown)
  }, [open])

  const handleClearComments = () => {
    if (window.confirm('Clear all comments?')) {
      onClearComments()
      setOpen(false)
    }
  }

  const handleClearChecks = () => {
    if (window.confirm('Clear all review checks?')) {
      onClearReviewedFiles()
      setOpen(false)
    }
  }

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="px-3 py-1 text-xs text-gray-400 hover:text-gray-200 border border-gray-700 rounded"
        aria-label="Clear"
      >
        Clear
      </button>
      {open && (
        <div className="absolute right-0 mt-1 w-40 bg-[#161b22] border border-gray-700 rounded-md shadow-lg z-10">
          <button
            type="button"
            onClick={handleClearComments}
            disabled={!hasComments}
            className="block w-full text-left px-3 py-2 text-sm text-gray-300 hover:bg-[#1c2128] disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Clear Comments
          </button>
          <button
            type="button"
            onClick={handleClearChecks}
            disabled={!hasReviewedFiles}
            className="block w-full text-left px-3 py-2 text-sm text-gray-300 hover:bg-[#1c2128] disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Clear Checks
          </button>
        </div>
      )}
    </div>
  )
}
