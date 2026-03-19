import { describe, it, expect } from 'vitest'
import { formatFileComments, formatAllComments, formatCommentPrefix } from './formatComments'
import type { Comment } from '../api/types'

const makeComment = (overrides: Partial<Comment> & { filePath: string; line: number; body: string }): Comment => ({
  id: 'c1',
  createdAt: '2026-01-01T00:00:00Z',
  ...overrides,
})

describe('formatCommentPrefix', () => {
  it('returns empty string when both are undefined', () => {
    expect(formatCommentPrefix()).toBe('')
  })

  it('returns category only when severity is undefined', () => {
    expect(formatCommentPrefix('MUST')).toBe('[MUST]')
  })

  it('returns severity only when category is undefined', () => {
    expect(formatCommentPrefix(undefined, 'Critical')).toBe('[Critical]')
  })

  it('returns both when both are set', () => {
    expect(formatCommentPrefix('MUST', 'Critical')).toBe('[MUST/Critical]')
  })
})

describe('formatFileComments', () => {
  it('returns empty string for empty array', () => {
    expect(formatFileComments('main.go', [])).toBe('')
  })

  it('formats single comment', () => {
    const comments = [makeComment({ filePath: 'main.go', line: 10, body: 'Fix this' })]
    expect(formatFileComments('main.go', comments)).toBe(
      '## main.go\n\n- **Line 10**: Fix this\n',
    )
  })

  it('formats file-level comment with File label', () => {
    const comments = [makeComment({ filePath: 'main.go', line: 0, body: 'General feedback' })]
    expect(formatFileComments('main.go', comments)).toBe(
      '## main.go\n\n- **File**: General feedback\n',
    )
  })

  it('sorts comments by line number', () => {
    const comments = [
      makeComment({ id: 'c2', filePath: 'main.go', line: 25, body: 'Second' }),
      makeComment({ id: 'c1', filePath: 'main.go', line: 10, body: 'First' }),
    ]
    const result = formatFileComments('main.go', comments)
    expect(result).toBe(
      '## main.go\n\n- **Line 10**: First\n- **Line 25**: Second\n',
    )
  })

  it('includes prefix when category and severity are set', () => {
    const comments = [
      makeComment({ filePath: 'main.go', line: 10, body: 'Fix this', reviewCategory: 'MUST', severity: 'Critical' }),
    ]
    expect(formatFileComments('main.go', comments)).toBe(
      '## main.go\n\n- **Line 10**: [MUST/Critical] Fix this\n',
    )
  })

  it('includes prefix with category only', () => {
    const comments = [
      makeComment({ filePath: 'main.go', line: 10, body: 'Consider', reviewCategory: 'IMO' }),
    ]
    expect(formatFileComments('main.go', comments)).toBe(
      '## main.go\n\n- **Line 10**: [IMO] Consider\n',
    )
  })
})

describe('formatAllComments', () => {
  it('returns empty string for empty array', () => {
    expect(formatAllComments([])).toBe('')
  })

  it('formats comments from a single file', () => {
    const comments = [
      makeComment({ filePath: 'main.go', line: 10, body: 'Fix this' }),
    ]
    expect(formatAllComments(comments)).toBe(
      '# Code Review Comments\n\n## main.go\n\n- **Line 10**: Fix this\n\n',
    )
  })

  it('groups and sorts comments by file path', () => {
    const comments = [
      makeComment({ id: 'c2', filePath: 'utils.ts', line: 5, body: 'Check this' }),
      makeComment({ id: 'c1', filePath: 'main.go', line: 10, body: 'Fix this' }),
    ]
    const result = formatAllComments(comments)
    expect(result).toBe(
      '# Code Review Comments\n\n' +
      '## main.go\n\n- **Line 10**: Fix this\n\n' +
      '## utils.ts\n\n- **Line 5**: Check this\n\n',
    )
  })

  it('sorts comments within each file by line number', () => {
    const comments = [
      makeComment({ id: 'c2', filePath: 'main.go', line: 25, body: 'Second' }),
      makeComment({ id: 'c1', filePath: 'main.go', line: 10, body: 'First' }),
    ]
    const result = formatAllComments(comments)
    expect(result).toContain('- **Line 10**: First\n- **Line 25**: Second\n')
  })
})
