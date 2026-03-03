import { useState } from 'react'
import type { TreeNode } from './FileTree'
import type { FileStatus } from '../../api/types'

interface DirectoryTreeProps {
  nodes: TreeNode[]
  selectedFile: string | null
  onSelectFile: (path: string) => void
}

const statusBadge: Record<FileStatus, { label: string; color: string }> = {
  added: { label: 'A', color: 'text-green-400' },
  deleted: { label: 'D', color: 'text-red-400' },
  modified: { label: 'M', color: 'text-yellow-400' },
  renamed: { label: 'R', color: 'text-blue-400' },
}

export function DirectoryTree({ nodes, selectedFile, onSelectFile }: DirectoryTreeProps) {
  return (
    <div className="text-sm">
      {nodes.map((node) => (
        <TreeItem
          key={node.path}
          node={node}
          depth={0}
          selectedFile={selectedFile}
          onSelectFile={onSelectFile}
        />
      ))}
    </div>
  )
}

function TreeItem({
  node,
  depth,
  selectedFile,
  onSelectFile,
}: {
  node: TreeNode
  depth: number
  selectedFile: string | null
  onSelectFile: (path: string) => void
}) {
  const [expanded, setExpanded] = useState(depth < 1 || !!node.hasChangedDescendant)

  if (node.isDirectory) {
    return (
      <div>
        <button
          type="button"
          onClick={() => setExpanded(!expanded)}
          className="w-full flex items-center gap-1 px-2 py-0.5 hover:bg-[#1c2128] text-left text-gray-300"
          style={{ paddingLeft: `${depth * 16 + 8}px` }}
        >
          <span className={`inline-block text-xs transition-transform ${expanded ? 'rotate-90' : ''}`}>&#9654;</span>
          <span className="font-mono text-xs">{node.name}</span>
        </button>
        {expanded && node.children.map((child) => (
          <TreeItem
            key={child.path}
            node={child}
            depth={depth + 1}
            selectedFile={selectedFile}
            onSelectFile={onSelectFile}
          />
        ))}
      </div>
    )
  }

  const isSelected = selectedFile === node.path
  const badge = node.status ? statusBadge[node.status] : null

  return (
    <button
      type="button"
      onClick={() => onSelectFile(node.path)}
      className={`w-full flex items-center gap-1 px-2 py-0.5 hover:bg-[#1c2128] text-left ${isSelected ? 'bg-[#1c2128]' : ''}`}
      style={{ paddingLeft: `${depth * 16 + 8}px` }}
    >
      <span className={`font-mono text-xs ${node.isChanged ? 'text-gray-200' : 'text-gray-400'}`}>
        {node.name}
      </span>
      {badge && (
        <span className={`text-xs font-bold ${badge.color} ml-auto`}>
          {badge.label}
        </span>
      )}
    </button>
  )
}
