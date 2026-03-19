import type { Comment, ReviewCategory, Severity } from '../api/types'

export function formatCommentPrefix(category?: ReviewCategory, severity?: Severity): string {
  if (!category && !severity) return ''
  if (!severity) return `[${category}]`
  if (!category) return `[${severity}]`
  return `[${category}/${severity}]`
}

export function formatFileComments(filePath: string, comments: Comment[]): string {
  if (comments.length === 0) return ''

  const sorted = [...comments].sort((a, b) => a.line - b.line)
  let result = `## ${filePath}\n\n`
  for (const c of sorted) {
    const prefix = formatCommentPrefix(c.reviewCategory, c.severity)
    if (c.line === 0) {
      result += prefix ? `- **File**: ${prefix}\n${c.body}\n` : `- **File**: ${c.body}\n`
    } else {
      result += prefix ? `- **Line ${c.line}**: ${prefix}\n${c.body}\n` : `- **Line ${c.line}**: ${c.body}\n`
    }
  }
  return result
}

export function formatAllComments(comments: Comment[]): string {
  if (comments.length === 0) return ''

  // Group by file
  const grouped = new Map<string, Comment[]>()
  for (const c of comments) {
    const arr = grouped.get(c.filePath)
    if (arr) {
      arr.push(c)
    } else {
      grouped.set(c.filePath, [c])
    }
  }

  // Sort file names
  const files = [...grouped.keys()].sort()

  let result = '# Code Review Comments\n\n'
  for (const file of files) {
    result += formatFileComments(file, grouped.get(file)!) + '\n'
  }
  return result
}
