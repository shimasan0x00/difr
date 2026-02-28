import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { HighlightedLine } from './SyntaxHighlight'

describe('HighlightedLine', () => {
  it('renders code content', async () => {
    render(<HighlightedLine code="const x = 1" language="typescript" />)

    await waitFor(() => {
      expect(screen.getByText(/const/)).toBeInTheDocument()
    })
  })

  it('produces syntax-highlighted tokens with spans', async () => {
    const { container } = render(
      <HighlightedLine code='console.log("hello")' language="typescript" />
    )

    await waitFor(() => {
      const spans = container.querySelectorAll('span[style]')
      expect(spans.length).toBeGreaterThan(0)
    })
  })

  it('renders plain text for unknown language', async () => {
    render(<HighlightedLine code="some text" language="unknown-lang-xyz" />)

    await waitFor(() => {
      expect(screen.getByText('some text')).toBeInTheDocument()
    })
  })

  it('renders without error for empty code string', async () => {
    const { container } = render(<HighlightedLine code="" language="typescript" />)

    await waitFor(() => {
      const span = container.querySelector('span')
      expect(span).toBeInTheDocument()
      expect(span?.textContent).toBe('')
    })
  })
})
