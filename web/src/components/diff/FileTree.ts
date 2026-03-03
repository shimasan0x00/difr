import type { FileStatus } from '../../api/types'

export interface TreeNode {
  name: string
  path: string
  isDirectory: boolean
  children: TreeNode[]
  isChanged?: boolean
  status?: FileStatus
  hasChangedDescendant?: boolean
}

export function buildFileTree(paths: string[], changedFiles: Map<string, FileStatus>): TreeNode[] {
  const root: TreeNode = { name: '', path: '', isDirectory: true, children: [] }

  for (const filePath of paths) {
    const parts = filePath.split('/')
    let current = root

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i]
      const isLast = i === parts.length - 1
      const currentPath = parts.slice(0, i + 1).join('/')

      if (isLast) {
        const status = changedFiles.get(filePath)
        current.children.push({
          name: part,
          path: currentPath,
          isDirectory: false,
          children: [],
          isChanged: status !== undefined ? true : undefined,
          status,
        })
      } else {
        let dir = current.children.find((c) => c.isDirectory && c.name === part)
        if (!dir) {
          dir = { name: part, path: currentPath, isDirectory: true, children: [] }
          current.children.push(dir)
        }
        current = dir
      }
    }
  }

  sortTree(root.children)
  markChangedAncestors(root.children)
  return root.children
}

function markChangedAncestors(nodes: TreeNode[]): boolean {
  let hasChanged = false
  for (const node of nodes) {
    if (node.isDirectory) {
      const descendantChanged = markChangedAncestors(node.children)
      node.hasChangedDescendant = descendantChanged
      if (descendantChanged) hasChanged = true
    } else if (node.isChanged) {
      hasChanged = true
    }
  }
  return hasChanged
}

function sortTree(nodes: TreeNode[]): void {
  nodes.sort((a, b) => {
    if (a.isDirectory !== b.isDirectory) return a.isDirectory ? -1 : 1
    return a.name.localeCompare(b.name)
  })
  for (const node of nodes) {
    if (node.isDirectory) sortTree(node.children)
  }
}
