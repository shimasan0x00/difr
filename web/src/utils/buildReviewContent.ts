import type { DiffFile } from '../api/types'

export function buildReviewContent(files: DiffFile[]): string {
  const MAX_SIZE = 800_000
  let content = 'Review the following code changes:\n\n'
  let truncated = false
  for (const file of files) {
    const path = file.newPath || file.oldPath
    let section = `--- ${path} (${file.status}) ---\n`
    for (const hunk of file.hunks) {
      if (hunk.header) section += `@@ ${hunk.header} @@\n`
      for (const line of hunk.lines) {
        const prefix = line.type === 'add' ? '+' : line.type === 'delete' ? '-' : ' '
        section += `${prefix}${line.content}`
      }
    }
    section += '\n'
    if (content.length + section.length > MAX_SIZE) { truncated = true; break }
    content += section
  }
  if (truncated) content += '\n(Diff truncated due to size limit.)\n'
  return content
}
