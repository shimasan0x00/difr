import { useState, useRef, useEffect } from 'react'

export function ExportButton() {
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

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="px-3 py-1 text-xs text-gray-400 hover:text-gray-200 border border-gray-700 rounded"
        aria-label="Export"
      >
        Export
      </button>
      {open && (
        <div className="absolute right-0 mt-1 w-36 bg-[#161b22] border border-gray-700 rounded-md shadow-lg z-10">
          <a
            href="/api/comments/export?format=markdown"
            className="block px-3 py-2 text-sm text-gray-300 hover:bg-[#1c2128]"
            aria-label="Markdown"
          >
            Markdown
          </a>
          <a
            href="/api/comments/export?format=json"
            className="block px-3 py-2 text-sm text-gray-300 hover:bg-[#1c2128]"
            aria-label="JSON"
          >
            JSON
          </a>
          <a
            href="/api/comments/export?format=csv"
            className="block px-3 py-2 text-sm text-gray-300 hover:bg-[#1c2128]"
            aria-label="CSV"
          >
            CSV
          </a>
          <a
            href="/api/comments/export?format=xlsx"
            className="block px-3 py-2 text-sm text-gray-300 hover:bg-[#1c2128]"
            aria-label="Excel"
          >
            Excel
          </a>
        </div>
      )}
    </div>
  )
}
