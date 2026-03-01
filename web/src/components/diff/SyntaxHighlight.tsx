import { memo, useEffect, useState } from 'react'
import { createHighlighter, type BundledLanguage, type Highlighter, type ThemedToken } from 'shiki'

let highlighterPromise: Promise<Highlighter> | null = null

function getHighlighter(): Promise<Highlighter> {
  if (!highlighterPromise) {
    highlighterPromise = createHighlighter({
      themes: ['github-dark'],
      langs: [
        'javascript', 'typescript', 'go', 'python', 'rust', 'java',
        'c', 'cpp', 'csharp', 'ruby', 'php', 'swift', 'kotlin',
        'html', 'css', 'json', 'yaml', 'toml', 'markdown',
        'bash', 'sql', 'dockerfile',
      ],
    })
  }
  return highlighterPromise
}

interface HighlightedLineProps {
  code: string
  language: string
}

interface TokenLine {
  tokens: ThemedToken[]
}

export const HighlightedLine = memo(function HighlightedLine({ code, language }: HighlightedLineProps) {
  const [tokenLine, setTokenLine] = useState<TokenLine | null>(null)

  useEffect(() => {
    let cancelled = false

    getHighlighter()
      .then((highlighter) => {
        if (cancelled) return
        const loadedLangs = highlighter.getLoadedLanguages()
        if (!loadedLangs.includes(language as never)) {
          setTokenLine(null)
          return
        }
        const result = highlighter.codeToTokens(code, {
          lang: language as BundledLanguage,
          theme: 'github-dark',
        })
        if (!cancelled && result.tokens.length > 0) {
          setTokenLine({ tokens: result.tokens[0] })
        }
      })
      .catch((err) => {
        if (!cancelled) {
          console.error('Shiki highlighter initialization failed:', err)
          setTokenLine(null)
        }
      })

    return () => {
      cancelled = true
    }
  }, [code, language])

  if (tokenLine) {
    return (
      <span>
        {tokenLine.tokens.map((token, i) => (
          <span key={i} style={{ color: token.color }}>
            {token.content}
          </span>
        ))}
      </span>
    )
  }

  return <span>{code}</span>
})
