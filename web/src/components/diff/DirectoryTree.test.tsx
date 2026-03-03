import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { DirectoryTree } from './DirectoryTree'
import type { TreeNode } from './FileTree'

function makeTree(): TreeNode[] {
  return [
    {
      name: 'src',
      path: 'src',
      isDirectory: true,
      children: [
        {
          name: 'pkg',
          path: 'src/pkg',
          isDirectory: true,
          children: [
            { name: 'util.go', path: 'src/pkg/util.go', isDirectory: false, children: [] },
          ],
        },
        { name: 'main.go', path: 'src/main.go', isDirectory: false, children: [], isChanged: true, status: 'modified' },
      ],
    },
    { name: 'README.md', path: 'README.md', isDirectory: false, children: [], isChanged: true, status: 'added' },
  ]
}

describe('DirectoryTree', () => {
  it('renders top-level nodes', () => {
    render(
      <DirectoryTree
        nodes={makeTree()}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    expect(screen.getByText('src')).toBeInTheDocument()
    expect(screen.getByText('README.md')).toBeInTheDocument()
  })

  it('shows first 2 levels expanded by default', () => {
    render(
      <DirectoryTree
        nodes={makeTree()}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    // Level 1: src is expanded, shows its children
    expect(screen.getByText('main.go')).toBeInTheDocument()
    expect(screen.getByText('pkg')).toBeInTheDocument()
    // Level 2 (depth=2): pkg is NOT expanded by default
    expect(screen.queryByText('util.go')).not.toBeInTheDocument()
  })

  it('toggles directory expansion on click', async () => {
    const user = userEvent.setup()
    render(
      <DirectoryTree
        nodes={makeTree()}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    // pkg is collapsed — util.go not visible
    expect(screen.queryByText('util.go')).not.toBeInTheDocument()

    // Click pkg to expand
    await user.click(screen.getByText('pkg'))
    expect(screen.getByText('util.go')).toBeInTheDocument()

    // Click pkg again to collapse
    await user.click(screen.getByText('pkg'))
    expect(screen.queryByText('util.go')).not.toBeInTheDocument()
  })

  it('calls onSelectFile when a file is clicked', async () => {
    const user = userEvent.setup()
    const onSelectFile = vi.fn()
    render(
      <DirectoryTree
        nodes={makeTree()}
        selectedFile={null}
        onSelectFile={onSelectFile}
      />
    )

    await user.click(screen.getByText('README.md'))

    expect(onSelectFile).toHaveBeenCalledWith('README.md')
  })

  it('shows status badge for changed files', () => {
    render(
      <DirectoryTree
        nodes={makeTree()}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    // README.md should have 'A' badge for added
    const addedBadge = screen.getByText('A')
    expect(addedBadge).toBeInTheDocument()

    // main.go should have 'M' badge for modified
    const modifiedBadge = screen.getByText('M')
    expect(modifiedBadge).toBeInTheDocument()
  })

  it('highlights selected file', () => {
    const { container } = render(
      <DirectoryTree
        nodes={makeTree()}
        selectedFile="src/main.go"
        onSelectFile={vi.fn()}
      />
    )

    const selectedItem = container.querySelector('.bg-\\[\\#1c2128\\]')
    expect(selectedItem).toBeInTheDocument()
  })

  it('auto-expands directories with hasChangedDescendant', () => {
    const nodes: TreeNode[] = [
      {
        name: 'deep',
        path: 'deep',
        isDirectory: true,
        hasChangedDescendant: true,
        children: [
          {
            name: 'nested',
            path: 'deep/nested',
            isDirectory: true,
            hasChangedDescendant: true,
            children: [
              { name: 'changed.go', path: 'deep/nested/changed.go', isDirectory: false, children: [], isChanged: true, status: 'modified' },
            ],
          },
        ],
      },
    ]

    render(
      <DirectoryTree
        nodes={nodes}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    // Both deep and nested should be expanded, showing the changed file
    expect(screen.getByText('changed.go')).toBeInTheDocument()
  })

  it('does not auto-expand directories without hasChangedDescendant', () => {
    const nodes: TreeNode[] = [
      {
        name: 'lib',
        path: 'lib',
        isDirectory: true,
        hasChangedDescendant: false,
        children: [
          {
            name: 'deep',
            path: 'lib/deep',
            isDirectory: true,
            hasChangedDescendant: false,
            children: [
              { name: 'file.go', path: 'lib/deep/file.go', isDirectory: false, children: [] },
            ],
          },
        ],
      },
    ]

    render(
      <DirectoryTree
        nodes={nodes}
        selectedFile={null}
        onSelectFile={vi.fn()}
      />
    )

    // lib is at depth 0 so expanded by default, but lib/deep is at depth 1 with no changed descendants
    expect(screen.getByText('deep')).toBeInTheDocument()
    expect(screen.queryByText('file.go')).not.toBeInTheDocument()
  })
})
