import { useState } from 'react'

export function ExportButton() {
  const [open, setOpen] = useState(false)

  return (
    <div className="relative">
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
        </div>
      )}
    </div>
  )
}
