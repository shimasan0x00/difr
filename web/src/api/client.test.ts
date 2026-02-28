import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { fetchDiff, fetchComments, createComment, updateComment, deleteComment } from './client'

let fetchSpy: ReturnType<typeof vi.spyOn>

beforeEach(() => {
  fetchSpy = vi.spyOn(globalThis, 'fetch')
})

afterEach(() => {
  fetchSpy.mockRestore()
})

function mockResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  } as Response
}

describe('fetchDiff', () => {
  it('returns parsed DiffResult on success', async () => {
    const data = { files: [], stats: { additions: 0, deletions: 0 } }
    fetchSpy.mockResolvedValue(mockResponse(data))

    const result = await fetchDiff()

    expect(fetchSpy).toHaveBeenCalledWith('/api/diff', { signal: undefined })
    expect(result).toEqual(data)
  })

  it('throws on non-OK response', async () => {
    fetchSpy.mockResolvedValue(mockResponse(null, 500))

    await expect(fetchDiff()).rejects.toThrow('Failed to fetch diff: 500')
  })
})

describe('fetchComments', () => {
  it('fetches all comments without filter', async () => {
    const comments = [{ id: 'c1', filePath: 'a.go', line: 1, body: 'test', createdAt: '' }]
    fetchSpy.mockResolvedValue(mockResponse(comments))

    const result = await fetchComments()

    expect(fetchSpy).toHaveBeenCalledWith('/api/comments')
    expect(result).toEqual(comments)
  })

  it('fetches comments with file filter', async () => {
    fetchSpy.mockResolvedValue(mockResponse([]))

    await fetchComments('src/main.go')

    expect(fetchSpy).toHaveBeenCalledWith('/api/comments?file=src%2Fmain.go')
  })

  it('throws on non-OK response', async () => {
    fetchSpy.mockResolvedValue(mockResponse(null, 404))

    await expect(fetchComments()).rejects.toThrow('Failed to fetch comments: 404')
  })
})

describe('createComment', () => {
  it('sends POST with correct body and returns created comment', async () => {
    const created = { id: 'c1', filePath: 'main.go', line: 10, body: 'fix', createdAt: '' }
    fetchSpy.mockResolvedValue(mockResponse(created, 201))

    const result = await createComment('main.go', 10, 'fix')

    expect(fetchSpy).toHaveBeenCalledWith('/api/comments', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ filePath: 'main.go', line: 10, body: 'fix' }),
    })
    expect(result).toEqual(created)
  })

  it('throws on non-OK response', async () => {
    fetchSpy.mockResolvedValue(mockResponse(null, 400))

    await expect(createComment('', 0, '')).rejects.toThrow('Failed to create comment: 400')
  })
})

describe('updateComment', () => {
  it('sends PUT with correct body and returns updated comment', async () => {
    const updated = { id: 'c1', filePath: 'main.go', line: 10, body: 'updated', createdAt: '' }
    fetchSpy.mockResolvedValue(mockResponse(updated))

    const result = await updateComment('c1', 'updated')

    expect(fetchSpy).toHaveBeenCalledWith('/api/comments/c1', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ body: 'updated' }),
    })
    expect(result).toEqual(updated)
  })

  it('throws on non-OK response', async () => {
    fetchSpy.mockResolvedValue(mockResponse(null, 404))

    await expect(updateComment('c1', 'x')).rejects.toThrow('Failed to update comment: 404')
  })
})

describe('extractErrorMessage (via API functions)', () => {
  it('extracts server error message from JSON response body', async () => {
    fetchSpy.mockResolvedValue({
      ok: false,
      status: 400,
      json: () => Promise.resolve({ error: 'filePath is required' }),
    } as Response)

    await expect(fetchDiff()).rejects.toThrow('filePath is required')
  })

  it('falls back to status code when response body is not JSON', async () => {
    fetchSpy.mockResolvedValue({
      ok: false,
      status: 502,
      json: () => Promise.reject(new Error('invalid json')),
    } as Response)

    await expect(fetchDiff()).rejects.toThrow('Failed to fetch diff: 502')
  })

  it('falls back to status code when error field is missing', async () => {
    fetchSpy.mockResolvedValue({
      ok: false,
      status: 403,
      json: () => Promise.resolve({ message: 'not the expected field' }),
    } as Response)

    await expect(fetchDiff()).rejects.toThrow('Failed to fetch diff: 403')
  })
})

describe('deleteComment', () => {
  it('sends DELETE request', async () => {
    fetchSpy.mockResolvedValue(mockResponse(null, 204))

    await deleteComment('c1')

    expect(fetchSpy).toHaveBeenCalledWith('/api/comments/c1', { method: 'DELETE' })
  })

  it('throws on non-OK response', async () => {
    fetchSpy.mockResolvedValue(mockResponse(null, 404))

    await expect(deleteComment('c99')).rejects.toThrow('Failed to delete comment: 404')
  })
})
