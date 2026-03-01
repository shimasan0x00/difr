import { useState } from 'react'
import type { Comment, DiffFile, FileStatus } from '../../api/types'

interface FileListPanelProps {
  files: DiffFile[]
  commentsByFile: Map<string, Comment[]>
  selectedFile: string | null
  onSelectFile: (path: string) => void
}

const statusColors: Record<FileStatus, string> = {
  added: 'text-green-400',
  deleted: 'text-red-400',
  modified: 'text-yellow-400',
  renamed: 'text-blue-400',
}

export function FileListPanel({ files, commentsByFile, selectedFile, onSelectFile }: FileListPanelProps) {
  const [expanded, setExpanded] = useState(true)

  return (
    <div className="border border-gray-700 rounded-md overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 bg-[#161b22] border-b border-gray-700">
        <button
          type="button"
          onClick={() => setExpanded(!expanded)}
          aria-expanded={expanded}
          aria-label="Toggle file list"
          className="flex items-center gap-2 text-sm text-gray-200 hover:text-white"
        >
          <span className={`inline-block transition-transform text-xs ${expanded ? 'rotate-90' : ''}`}>&#9654;</span>
          Files ({files.length})
        </button>
      </div>
      {expanded && (
        <div className="divide-y divide-gray-800">
          {files.map((file) => {
            const path = file.newPath && file.newPath !== '/dev/null' ? file.newPath : file.oldPath
            const commentCount = commentsByFile.get(path)?.length ?? 0
            const isSelected = selectedFile === path

            return (
              <button
                key={path}
                type="button"
                onClick={() => onSelectFile(path)}
                className={`w-full flex items-center justify-between px-4 py-1.5 text-sm hover:bg-[#1c2128] text-left ${isSelected ? 'bg-[#1c2128] border-l-2 border-blue-500' : ''}`}
              >
                <div className="flex items-center gap-2 min-w-0">
                  <span className={`shrink-0 text-xs font-bold ${statusColors[file.status]}`}>
                    {file.status[0].toUpperCase()}
                  </span>
                  <span className="text-gray-300 font-mono text-xs truncate">{path}</span>
                </div>
                <div className="flex items-center gap-2 shrink-0 text-xs">
                  {commentCount > 0 && (
                    <span className="text-gray-400 bg-gray-700 rounded-full px-1.5 py-0.5">{commentCount}</span>
                  )}
                  {file.stats.additions > 0 && (
                    <span className="text-green-400">+{file.stats.additions}</span>
                  )}
                  {file.stats.deletions > 0 && (
                    <span className="text-red-400">-{file.stats.deletions}</span>
                  )}
                </div>
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
