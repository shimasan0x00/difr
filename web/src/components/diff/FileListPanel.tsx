import { useMemo, useState } from 'react'
import type { Comment, DiffFile, FileStatus } from '../../api/types'
import { buildFileTree } from './FileTree'
import { DirectoryTree } from './DirectoryTree'

type SidebarTab = 'changed' | 'all'

interface FileListPanelProps {
  files: DiffFile[]
  trackedFiles?: string[]
  commentsByFile: Map<string, Comment[]>
  selectedFile: string | null
  onSelectFile: (path: string) => void
  expanded?: boolean
  onToggle?: () => void
  activeTab?: SidebarTab
  onTabChange?: (tab: SidebarTab) => void
  reviewedFiles?: Set<string>
  onToggleReviewed?: (path: string) => void
}

const statusColors: Record<FileStatus, string> = {
  added: 'text-green-400',
  deleted: 'text-red-400',
  modified: 'text-yellow-400',
  renamed: 'text-blue-400',
}

export function FileListPanel({ files, trackedFiles = [], commentsByFile, selectedFile, onSelectFile, expanded: controlledExpanded, onToggle, activeTab = 'changed', onTabChange, reviewedFiles, onToggleReviewed }: FileListPanelProps) {
  const [internalExpanded, setInternalExpanded] = useState(true)
  const expanded = controlledExpanded ?? internalExpanded
  const toggleExpanded = onToggle ?? (() => setInternalExpanded(!internalExpanded))

  const hasTrackedFiles = trackedFiles.length > 0

  const changedFilesMap = useMemo(() => {
    const map = new Map<string, FileStatus>()
    for (const file of files) {
      if (file.newPath && file.newPath !== '/dev/null') map.set(file.newPath, file.status)
      if (file.oldPath && file.oldPath !== '/dev/null') map.set(file.oldPath, file.status)
    }
    return map
  }, [files])

  const treeNodes = useMemo(() => {
    if (!hasTrackedFiles) return []
    return buildFileTree(trackedFiles, changedFilesMap)
  }, [trackedFiles, changedFilesMap, hasTrackedFiles])

  return (
    <div className="border-b border-gray-700 overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 bg-[#161b22] border-b border-gray-700">
        <button
          type="button"
          onClick={toggleExpanded}
          aria-expanded={expanded}
          aria-label="Toggle file list"
          className="flex items-center gap-2 text-sm text-gray-200 hover:text-white"
        >
          <span className={`inline-block transition-transform text-xs ${expanded ? 'rotate-90' : ''}`}>&#9654;</span>
          Files ({files.length})
        </button>
      </div>
      {expanded && hasTrackedFiles && (
        <div className="flex border-b border-gray-700 bg-[#161b22]" role="tablist">
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'changed'}
            onClick={() => onTabChange?.('changed')}
            className={`flex-1 px-3 py-1.5 text-xs text-center ${activeTab === 'changed' ? 'text-white border-b-2 border-blue-500' : 'text-gray-400 hover:text-gray-200'}`}
          >
            Changed
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'all'}
            onClick={() => onTabChange?.('all')}
            className={`flex-1 px-3 py-1.5 text-xs text-center ${activeTab === 'all' ? 'text-white border-b-2 border-blue-500' : 'text-gray-400 hover:text-gray-200'}`}
          >
            All Files
          </button>
        </div>
      )}
      {expanded && activeTab === 'changed' && (
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
                  {onToggleReviewed && (
                    <input
                      type="checkbox"
                      checked={reviewedFiles?.has(path) ?? false}
                      onChange={() => onToggleReviewed(path)}
                      onClick={(e) => e.stopPropagation()}
                      aria-label={`Mark ${path} as reviewed`}
                      className="shrink-0"
                    />
                  )}
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
      {expanded && activeTab === 'all' && hasTrackedFiles && (
        <DirectoryTree
          nodes={treeNodes}
          selectedFile={selectedFile}
          onSelectFile={onSelectFile}
        />
      )}
    </div>
  )
}
