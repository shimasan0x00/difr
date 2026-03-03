import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { CopyAllButton } from './CopyAllButton'
import type { Comment } from '../../api/types'

describe('CopyAllButton', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders Copy All button', () => {
    const comments: Comment[] = [
      { id: 'c1', filePath: 'main.go', line: 10, body: 'Fix this', createdAt: '2026-01-01T00:00:00Z' },
    ]
    render(<CopyAllButton comments={comments} />)
    expect(screen.getByRole('button', { name: 'Copy all comments' })).toBeInTheDocument()
    expect(screen.getByText('Copy All')).toBeInTheDocument()
  })

  it('copies formatted comments to clipboard on click', async () => {
    const user = userEvent.setup()
    const writeText = vi.fn().mockResolvedValue(undefined)
    vi.spyOn(navigator.clipboard, 'writeText').mockImplementation(writeText)

    const comments: Comment[] = [
      { id: 'c1', filePath: 'main.go', line: 10, body: 'Fix this', createdAt: '2026-01-01T00:00:00Z' },
      { id: 'c2', filePath: 'utils.ts', line: 5, body: 'Check this', createdAt: '2026-01-01T00:00:00Z' },
    ]
    render(<CopyAllButton comments={comments} />)

    await user.click(screen.getByRole('button', { name: 'Copy all comments' }))

    expect(writeText).toHaveBeenCalledWith(
      '# Code Review Comments\n\n' +
      '## main.go\n\n- **Line 10**: Fix this\n\n' +
      '## utils.ts\n\n- **Line 5**: Check this\n\n',
    )
  })

  it('shows Copied! feedback after click', async () => {
    const user = userEvent.setup()
    vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue(undefined)

    const comments: Comment[] = [
      { id: 'c1', filePath: 'main.go', line: 10, body: 'Fix this', createdAt: '2026-01-01T00:00:00Z' },
    ]
    render(<CopyAllButton comments={comments} />)

    await user.click(screen.getByRole('button', { name: 'Copy all comments' }))
    expect(screen.getByText('Copied!')).toBeInTheDocument()
  })
})
