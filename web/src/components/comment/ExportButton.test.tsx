import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect } from 'vitest'
import { ExportButton } from './ExportButton'

describe('ExportButton', () => {
  it('renders Export button', () => {
    render(<ExportButton />)

    expect(screen.getByRole('button', { name: /export/i })).toBeInTheDocument()
  })

  it('shows dropdown with format options when clicked', async () => {
    const user = userEvent.setup()
    render(<ExportButton />)

    await user.click(screen.getByRole('button', { name: /export/i }))

    expect(screen.getByRole('link', { name: /markdown/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /json/i })).toBeInTheDocument()
  })

  it('has correct download links for each format', async () => {
    const user = userEvent.setup()
    render(<ExportButton />)

    await user.click(screen.getByRole('button', { name: /export/i }))

    const markdownLink = screen.getByRole('link', { name: /markdown/i })
    const jsonLink = screen.getByRole('link', { name: /json/i })

    expect(markdownLink).toHaveAttribute('href', '/api/comments/export?format=markdown')
    expect(jsonLink).toHaveAttribute('href', '/api/comments/export?format=json')
  })

  it('hides dropdown when clicking Export again', async () => {
    const user = userEvent.setup()
    render(<ExportButton />)

    await user.click(screen.getByRole('button', { name: /export/i }))
    expect(screen.getByRole('link', { name: /markdown/i })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /export/i }))
    expect(screen.queryByRole('link', { name: /markdown/i })).not.toBeInTheDocument()
  })
})
