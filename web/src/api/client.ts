import type { Comment, DiffResult } from './types'

async function extractErrorMessage(res: Response, fallback: string): Promise<string> {
  try {
    const body = await res.json()
    if (body && typeof body.error === 'string') return body.error
  } catch {
    // Response body is not JSON
  }
  return `${fallback}: ${res.status}`
}

export async function fetchDiff(signal?: AbortSignal): Promise<DiffResult> {
  const res = await fetch('/api/diff', { signal })
  if (!res.ok) {
    throw new Error(await extractErrorMessage(res, 'Failed to fetch diff'))
  }
  return res.json()
}

export async function fetchViewMode(signal?: AbortSignal): Promise<string> {
  const res = await fetch('/api/diff/mode', { signal })
  if (!res.ok) throw new Error(await extractErrorMessage(res, 'Failed to fetch view mode'))
  const data: { mode: string } = await res.json()
  return data.mode
}

export async function fetchComments(filePath?: string): Promise<Comment[]> {
  const params = filePath ? `?file=${encodeURIComponent(filePath)}` : ''
  const res = await fetch(`/api/comments${params}`)
  if (!res.ok) throw new Error(await extractErrorMessage(res, 'Failed to fetch comments'))
  return res.json()
}

export async function createComment(filePath: string, line: number, body: string): Promise<Comment> {
  const res = await fetch('/api/comments', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ filePath, line, body }),
  })
  if (!res.ok) throw new Error(await extractErrorMessage(res, 'Failed to create comment'))
  return res.json()
}

export async function updateComment(id: string, body: string): Promise<Comment> {
  const res = await fetch(`/api/comments/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ body }),
  })
  if (!res.ok) throw new Error(await extractErrorMessage(res, 'Failed to update comment'))
  return res.json()
}

export async function deleteComment(id: string): Promise<void> {
  const res = await fetch(`/api/comments/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(await extractErrorMessage(res, 'Failed to delete comment'))
}
