import { describe, it, expect } from 'vitest'
import { buildFileTree } from './FileTree'
import type { FileStatus } from '../../api/types'

describe('buildFileTree', () => {
  it('builds a flat list for root-level files', () => {
    const tree = buildFileTree(['main.go', 'utils.go'], new Map())

    expect(tree).toHaveLength(2)
    expect(tree[0].name).toBe('main.go')
    expect(tree[0].isDirectory).toBe(false)
    expect(tree[0].path).toBe('main.go')
    expect(tree[1].name).toBe('utils.go')
  })

  it('builds nested directory structure', () => {
    const tree = buildFileTree(['src/main.go', 'src/pkg/util.go'], new Map())

    expect(tree).toHaveLength(1)
    const src = tree[0]
    expect(src.name).toBe('src')
    expect(src.isDirectory).toBe(true)
    expect(src.children).toHaveLength(2)
    // Directory first, then file
    expect(src.children[0].name).toBe('pkg')
    expect(src.children[0].isDirectory).toBe(true)
    expect(src.children[1].name).toBe('main.go')
    expect(src.children[1].isDirectory).toBe(false)
  })

  it('sorts directories before files, then alphabetically', () => {
    const tree = buildFileTree([
      'b.go',
      'a/x.go',
      'a.go',
      'c/y.go',
    ], new Map())

    // Directories first: a/, c/, then files: a.go, b.go
    expect(tree.map((n) => n.name)).toEqual(['a', 'c', 'a.go', 'b.go'])
  })

  it('marks changed files with status', () => {
    const changedFiles = new Map<string, FileStatus>([
      ['src/main.go', 'modified'],
      ['README.md', 'added'],
    ])
    const tree = buildFileTree(['src/main.go', 'src/utils.go', 'README.md'], changedFiles)

    const readme = tree.find((n) => n.name === 'README.md')
    expect(readme?.isChanged).toBe(true)
    expect(readme?.status).toBe('added')

    const src = tree.find((n) => n.name === 'src')!
    const mainGo = src.children.find((n) => n.name === 'main.go')
    expect(mainGo?.isChanged).toBe(true)
    expect(mainGo?.status).toBe('modified')

    const utilsGo = src.children.find((n) => n.name === 'utils.go')
    expect(utilsGo?.isChanged).toBeFalsy()
  })

  it('returns empty array for empty input', () => {
    const tree = buildFileTree([], new Map())
    expect(tree).toEqual([])
  })

  it('handles deeply nested paths', () => {
    const tree = buildFileTree(['a/b/c/d.go'], new Map())

    expect(tree).toHaveLength(1)
    expect(tree[0].name).toBe('a')
    expect(tree[0].children[0].name).toBe('b')
    expect(tree[0].children[0].children[0].name).toBe('c')
    expect(tree[0].children[0].children[0].children[0].name).toBe('d.go')
  })

  describe('hasChangedDescendant', () => {
    it('marks directories containing changed files', () => {
      const changedFiles = new Map<string, FileStatus>([['src/main.go', 'modified']])
      const tree = buildFileTree(['src/main.go', 'src/utils.go', 'README.md'], changedFiles)

      const src = tree.find((n) => n.name === 'src')!
      expect(src.hasChangedDescendant).toBe(true)
    })

    it('marks all ancestor directories for deeply nested changed files', () => {
      const changedFiles = new Map<string, FileStatus>([['a/b/c/d.go', 'added']])
      const tree = buildFileTree(['a/b/c/d.go', 'a/b/other.go'], changedFiles)

      expect(tree[0].hasChangedDescendant).toBe(true) // a
      expect(tree[0].children[0].hasChangedDescendant).toBe(true) // a/b
      expect(tree[0].children[0].children[0].hasChangedDescendant).toBe(true) // a/b/c
    })

    it('does not mark directories without changed files', () => {
      const changedFiles = new Map<string, FileStatus>([['src/main.go', 'modified']])
      const tree = buildFileTree(['src/main.go', 'lib/utils.go'], changedFiles)

      const lib = tree.find((n) => n.name === 'lib')!
      expect(lib.hasChangedDescendant).toBe(false)
    })
  })
})
