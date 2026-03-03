import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ClearButton } from './ClearButton'

describe('ClearButton', () => {
  const defaultProps = {
    onClearComments: vi.fn(),
    onClearReviewedFiles: vi.fn(),
    hasComments: true,
    hasReviewedFiles: true,
  }

  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('shows dropdown when clicked', async () => {
    const user = userEvent.setup()
    render(<ClearButton {...defaultProps} />)

    await user.click(screen.getByRole('button', { name: 'Clear' }))

    expect(screen.getByText('Clear Comments')).toBeInTheDocument()
    expect(screen.getByText('Clear Checks')).toBeInTheDocument()
  })

  it('calls onClearComments after confirm', async () => {
    const user = userEvent.setup()
    const onClear = vi.fn()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    render(<ClearButton {...defaultProps} onClearComments={onClear} />)

    await user.click(screen.getByRole('button', { name: 'Clear' }))
    await user.click(screen.getByText('Clear Comments'))

    expect(window.confirm).toHaveBeenCalledWith('Clear all comments?')
    expect(onClear).toHaveBeenCalledTimes(1)
  })

  it('does not call onClearComments when confirm is cancelled', async () => {
    const user = userEvent.setup()
    const onClear = vi.fn()
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    render(<ClearButton {...defaultProps} onClearComments={onClear} />)

    await user.click(screen.getByRole('button', { name: 'Clear' }))
    await user.click(screen.getByText('Clear Comments'))

    expect(onClear).not.toHaveBeenCalled()
  })

  it('calls onClearReviewedFiles after confirm', async () => {
    const user = userEvent.setup()
    const onClear = vi.fn()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    render(<ClearButton {...defaultProps} onClearReviewedFiles={onClear} />)

    await user.click(screen.getByRole('button', { name: 'Clear' }))
    await user.click(screen.getByText('Clear Checks'))

    expect(window.confirm).toHaveBeenCalledWith('Clear all review checks?')
    expect(onClear).toHaveBeenCalledTimes(1)
  })

  it('disables Clear Comments when hasComments is false', async () => {
    const user = userEvent.setup()
    render(<ClearButton {...defaultProps} hasComments={false} />)

    await user.click(screen.getByRole('button', { name: 'Clear' }))

    expect(screen.getByText('Clear Comments')).toBeDisabled()
  })

  it('disables Clear Checks when hasReviewedFiles is false', async () => {
    const user = userEvent.setup()
    render(<ClearButton {...defaultProps} hasReviewedFiles={false} />)

    await user.click(screen.getByRole('button', { name: 'Clear' }))

    expect(screen.getByText('Clear Checks')).toBeDisabled()
  })
})
