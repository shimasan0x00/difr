import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { FileViewer } from './FileViewer'
import type { FileContent } from '../../api/types'

describe('FileViewer', () => {
  it('renders file content with line numbers', () => {
    const fc: FileContent = {
      path: 'main.go',
      content: 'package main\n\nfunc main() {}\n',
      size: 30,
    }

    render(<FileViewer file={fc} />)

    expect(screen.getByText('main.go')).toBeInTheDocument()
    expect(screen.getByText('package main')).toBeInTheDocument()
    expect(screen.getByText('func main() {}')).toBeInTheDocument()
    // Line numbers
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('shows binary file message', () => {
    const fc: FileContent = {
      path: 'image.png',
      content: '',
      isBinary: true,
      size: 1024,
    }

    render(<FileViewer file={fc} />)

    expect(screen.getByText('image.png')).toBeInTheDocument()
    expect(screen.getByText('Binary file (1.0 KB)')).toBeInTheDocument()
  })

  it('shows truncated file message', () => {
    const fc: FileContent = {
      path: 'large.txt',
      content: '',
      isTruncated: true,
      size: 6 * 1024 * 1024,
    }

    render(<FileViewer file={fc} />)

    expect(screen.getByText('large.txt')).toBeInTheDocument()
    expect(screen.getByText('File too large to display (6.0 MB)')).toBeInTheDocument()
  })

  it('shows empty file message for empty content', () => {
    const fc: FileContent = {
      path: 'empty.txt',
      content: '',
      size: 0,
    }

    render(<FileViewer file={fc} />)

    expect(screen.getByText('empty.txt')).toBeInTheDocument()
    expect(screen.getByText('Empty file')).toBeInTheDocument()
  })

  it('renders file size in header', () => {
    const fc: FileContent = {
      path: 'main.go',
      content: 'package main\n',
      size: 13,
    }

    render(<FileViewer file={fc} />)

    expect(screen.getByText('13 B')).toBeInTheDocument()
  })
})
