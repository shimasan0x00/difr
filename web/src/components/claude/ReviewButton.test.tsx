import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { ReviewButton } from './ReviewButton'

describe('ReviewButton', () => {
  it('renders with "Auto Review" label', () => {
    render(<ReviewButton onClick={vi.fn()} loading={false} />)

    expect(screen.getByRole('button', { name: /auto review/i })).toBeInTheDocument()
  })

  it('calls onClick when clicked', async () => {
    const user = userEvent.setup()
    const onClick = vi.fn()
    render(<ReviewButton onClick={onClick} loading={false} />)

    await user.click(screen.getByRole('button', { name: /auto review/i }))

    expect(onClick).toHaveBeenCalledOnce()
  })

  it('shows "Reviewing..." and is disabled when loading', () => {
    render(<ReviewButton onClick={vi.fn()} loading={true} />)

    const button = screen.getByRole('button', { name: /reviewing/i })
    expect(button).toBeDisabled()
  })
})
