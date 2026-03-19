import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { useCommentStore } from './commentStore'
import type { Comment } from '../api/types'

// Mock the API client
vi.mock('../api/client', () => ({
  createComment: vi.fn(),
  updateComment: vi.fn(),
  deleteComment: vi.fn(),
  fetchComments: vi.fn(),
  deleteAllComments: vi.fn(),
}))

import * as api from '../api/client'

const mockComment: Comment = {
  id: 'c1',
  filePath: 'main.go',
  line: 10,
  body: 'needs fix',
  createdAt: '2026-01-01T00:00:00Z',
}

function resetStore() {
  useCommentStore.setState({
    comments: [],
    loading: false,
    saving: false,
    error: null,
  })
}

describe('commentStore', () => {
  beforeEach(() => {
    resetStore()
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('addComment', () => {
    it('appends created comment to list and sets saving state', async () => {
      vi.mocked(api.createComment).mockResolvedValue(mockComment)

      await useCommentStore.getState().addComment('main.go', 10, 'needs fix')

      const state = useCommentStore.getState()
      expect(state.comments).toHaveLength(1)
      expect(state.comments[0]).toEqual(mockComment)
      expect(state.saving).toBe(false)
      expect(state.error).toBeNull()
    })

    it('sets error on API failure', async () => {
      vi.mocked(api.createComment).mockRejectedValue(new Error('network error'))

      await useCommentStore.getState().addComment('main.go', 10, 'fix')

      const state = useCommentStore.getState()
      expect(state.comments).toHaveLength(0)
      expect(state.saving).toBe(false)
      expect(state.error).toBe('network error')
    })

    it('passes reviewCategory and severity to API', async () => {
      vi.mocked(api.createComment).mockResolvedValue({ ...mockComment, reviewCategory: 'MUST', severity: 'Critical' })

      await useCommentStore.getState().addComment('main.go', 10, 'needs fix', 'MUST', 'Critical')

      expect(api.createComment).toHaveBeenCalledWith('main.go', 10, 'needs fix', 'MUST', 'Critical')
      const state = useCommentStore.getState()
      expect(state.comments[0].reviewCategory).toBe('MUST')
      expect(state.comments[0].severity).toBe('Critical')
    })

    it('handles non-Error rejection with fallback message', async () => {
      vi.mocked(api.createComment).mockRejectedValue('string error')

      await useCommentStore.getState().addComment('main.go', 10, 'fix')

      expect(useCommentStore.getState().error).toBe('Failed to add comment')
    })

    it('prevents duplicate submission while saving', async () => {
      let resolveFirst: (value: Comment) => void
      const firstCall = new Promise<Comment>((resolve) => {
        resolveFirst = resolve
      })
      vi.mocked(api.createComment).mockReturnValueOnce(firstCall)

      // Start first addComment (sets saving=true)
      const firstPromise = useCommentStore.getState().addComment('main.go', 10, 'first')

      // Second addComment should be skipped because saving=true
      await useCommentStore.getState().addComment('main.go', 20, 'second')

      // Resolve first call
      resolveFirst!(mockComment)
      await firstPromise

      expect(api.createComment).toHaveBeenCalledTimes(1)
    })
  })

  describe('updateComment', () => {
    it('replaces comment in list with updated version', async () => {
      const updated = { ...mockComment, body: 'updated body' }
      useCommentStore.setState({ comments: [mockComment] })
      vi.mocked(api.updateComment).mockResolvedValue(updated)

      await useCommentStore.getState().updateComment('c1', 'updated body')

      const state = useCommentStore.getState()
      expect(state.comments).toHaveLength(1)
      expect(state.comments[0].body).toBe('updated body')
      expect(state.saving).toBe(false)
    })

    it('passes reviewCategory and severity to API', async () => {
      const updated = { ...mockComment, body: 'updated', reviewCategory: 'IMO' as const, severity: 'High' as const }
      useCommentStore.setState({ comments: [mockComment] })
      vi.mocked(api.updateComment).mockResolvedValue(updated)

      await useCommentStore.getState().updateComment('c1', 'updated', 'IMO', 'High')

      expect(api.updateComment).toHaveBeenCalledWith('c1', 'updated', 'IMO', 'High')
      expect(useCommentStore.getState().comments[0].reviewCategory).toBe('IMO')
    })

    it('sets error on API failure', async () => {
      useCommentStore.setState({ comments: [mockComment] })
      vi.mocked(api.updateComment).mockRejectedValue(new Error('not found'))

      await useCommentStore.getState().updateComment('c1', 'new')

      expect(useCommentStore.getState().error).toBe('not found')
      // Original comment should remain unchanged
      expect(useCommentStore.getState().comments[0].body).toBe('needs fix')
    })
  })

  describe('removeComment', () => {
    it('removes comment from list', async () => {
      useCommentStore.setState({ comments: [mockComment] })
      vi.mocked(api.deleteComment).mockResolvedValue(undefined)

      await useCommentStore.getState().removeComment('c1')

      expect(useCommentStore.getState().comments).toHaveLength(0)
      expect(useCommentStore.getState().saving).toBe(false)
    })

    it('sets error on API failure', async () => {
      useCommentStore.setState({ comments: [mockComment] })
      vi.mocked(api.deleteComment).mockRejectedValue(new Error('server error'))

      await useCommentStore.getState().removeComment('c1')

      expect(useCommentStore.getState().error).toBe('server error')
      // Comment should still be in list since delete failed
      expect(useCommentStore.getState().comments).toHaveLength(1)
    })
  })

  describe('clearAll', () => {
    it('clears all comments optimistically', async () => {
      useCommentStore.setState({ comments: [mockComment] })
      vi.mocked(api.deleteAllComments).mockResolvedValue(undefined)

      await useCommentStore.getState().clearAll()

      expect(useCommentStore.getState().comments).toHaveLength(0)
      expect(api.deleteAllComments).toHaveBeenCalled()
    })

    it('rolls back on API failure', async () => {
      useCommentStore.setState({ comments: [mockComment] })
      vi.mocked(api.deleteAllComments).mockRejectedValue(new Error('fail'))

      await useCommentStore.getState().clearAll()

      expect(useCommentStore.getState().comments).toHaveLength(1)
    })
  })

  describe('loadComments', () => {
    it('loads comments and sets loading state', async () => {
      vi.mocked(api.fetchComments).mockResolvedValue([mockComment])

      await useCommentStore.getState().loadComments()

      const state = useCommentStore.getState()
      expect(state.comments).toEqual([mockComment])
      expect(state.loading).toBe(false)
      expect(state.error).toBeNull()
    })

    it('passes filePath filter to API', async () => {
      vi.mocked(api.fetchComments).mockResolvedValue([])

      await useCommentStore.getState().loadComments('main.go')

      expect(api.fetchComments).toHaveBeenCalledWith('main.go')
    })

    it('sets error on API failure', async () => {
      vi.mocked(api.fetchComments).mockRejectedValue(new Error('fetch failed'))

      await useCommentStore.getState().loadComments()

      const state = useCommentStore.getState()
      expect(state.error).toBe('fetch failed')
      expect(state.loading).toBe(false)
    })
  })
})
